package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/kavos113/quickctf/ctf-manager/domain"
	managerPb "github.com/kavos113/quickctf/gen/go/api/manager/v1"
	runnerPb "github.com/kavos113/quickctf/gen/go/api/runner/v1"
)

type ManagerService struct {
	managerPb.UnimplementedRunnerServiceServer
	runners    []*RunnerClient
	repo       domain.InstanceRepository
	mu         sync.RWMutex
	nextRunner int
}

type RunnerClient struct {
	URL        string
	Client     runnerPb.RunnerServiceClient
	Connection *grpc.ClientConn
	Active     bool
}

func NewManagerService(runnerURLs []string, repo domain.InstanceRepository) (*ManagerService, error) {
	runners := make([]*RunnerClient, 0, len(runnerURLs))

	for _, url := range runnerURLs {
		conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("Failed to connect to runner %s: %v", url, err)
			continue
		}

		client := runnerPb.NewRunnerServiceClient(conn)
		runners = append(runners, &RunnerClient{
			URL:        url,
			Client:     client,
			Connection: conn,
			Active:     true,
		})
		log.Printf("Connected to runner: %s", url)
	}

	if len(runners) == 0 {
		return nil, fmt.Errorf("no runners available")
	}

	return &ManagerService{
		runners: runners,
		repo:    repo,
	}, nil
}

func (s *ManagerService) selectRunner() *RunnerClient {
	s.mu.Lock()
	defer s.mu.Unlock()

	// ラウンドロビン方式でrunnerを選択
	startIdx := s.nextRunner
	for i := 0; i < len(s.runners); i++ {
		idx := (startIdx + i) % len(s.runners)
		if s.runners[idx].Active {
			s.nextRunner = (idx + 1) % len(s.runners)
			return s.runners[idx]
		}
	}

	return nil
}

func (s *ManagerService) getRunnerByURL(url string) *RunnerClient {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, runner := range s.runners {
		if runner.URL == url {
			return runner
		}
	}
	return nil
}

func (s *ManagerService) StartInstance(ctx context.Context, req *managerPb.StartInstanceRequest) (*managerPb.StartInstanceResponse, error) {
	_, err := s.repo.FindByID(ctx, req.InstanceId)
	if err == nil {
		return &managerPb.StartInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("instance %s already exists", req.InstanceId),
		}, nil
	}
	if err != domain.ErrInstanceNotFound {
		return &managerPb.StartInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("failed to check instance: %v", err),
		}, nil
	}

	runner := s.selectRunner()
	if runner == nil {
		return &managerPb.StartInstanceResponse{
			Status:       "failed",
			ErrorMessage: "no available runners",
		}, nil
	}

	runnerReq := &runnerPb.StartInstanceRequest{
		ImageTag:   req.ImageTag,
		InstanceId: req.InstanceId,
	}

	resp, err := runner.Client.StartInstance(ctx, runnerReq)
	if err != nil {
		return &managerPb.StartInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("failed to start instance on runner: %v", err),
		}, nil
	}

	if resp.Status == "success" {
		now := time.Now()
		instance := &domain.Instance{
			InstanceID: req.InstanceId,
			ImageTag:   req.ImageTag,
			RunnerURL:  runner.URL,
			Host:       resp.ConnectionInfo.Host,
			Port:       resp.ConnectionInfo.Port,
			State:      domain.StateRunning,
			TTL:        time.Duration(req.TtlSeconds) * time.Second,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		if err := s.repo.Create(ctx, instance); err != nil {
			log.Printf("Failed to save instance %s: %v", req.InstanceId, err)
			// DBへの保存に失敗してもrunnerで起動しているので、失敗として扱わない
		} else {
			log.Printf("Instance %s started on runner %s", req.InstanceId, runner.URL)
		}
	}

	var connInfo *managerPb.ConnectionInfo
	if resp.ConnectionInfo != nil {
		connInfo = &managerPb.ConnectionInfo{
			Host: resp.ConnectionInfo.Host,
			Port: resp.ConnectionInfo.Port,
		}
	}

	return &managerPb.StartInstanceResponse{
		Status:         resp.Status,
		ErrorMessage:   resp.ErrorMessage,
		ConnectionInfo: connInfo,
	}, nil
}

func (s *ManagerService) StopInstance(ctx context.Context, req *managerPb.StopInstanceRequest) (*managerPb.StopInstanceResponse, error) {
	instance, err := s.repo.FindByID(ctx, req.InstanceId)
	if err == domain.ErrInstanceNotFound {
		return &managerPb.StopInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("instance %s not found", req.InstanceId),
		}, nil
	}
	if err != nil {
		return &managerPb.StopInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("failed to get instance: %v", err),
		}, nil
	}

	runner := s.getRunnerByURL(instance.RunnerURL)
	if runner == nil {
		return &managerPb.StopInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("runner %s not found", instance.RunnerURL),
		}, nil
	}

	runnerReq := &runnerPb.StopInstanceRequest{
		InstanceId: req.InstanceId,
	}

	resp, err := runner.Client.StopInstance(ctx, runnerReq)
	if err != nil {
		return &managerPb.StopInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("failed to stop instance on runner: %v", err),
		}, nil
	}

	if resp.Status == "success" {
		instance.UpdateState(domain.StateStopped)
		if err := s.repo.Update(ctx, instance); err != nil {
			log.Printf("Failed to update instance %s state: %v", req.InstanceId, err)
		} else {
			log.Printf("Instance %s stopped on runner %s", req.InstanceId, runner.URL)
		}
	}

	return &managerPb.StopInstanceResponse{
		Status:       resp.Status,
		ErrorMessage: resp.ErrorMessage,
	}, nil
}

func (s *ManagerService) DestroyInstance(ctx context.Context, req *managerPb.DestroyInstanceRequest) (*managerPb.DestroyInstanceResponse, error) {
	instance, err := s.repo.FindByID(ctx, req.InstanceId)
	if err == domain.ErrInstanceNotFound {
		return &managerPb.DestroyInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("instance %s not found", req.InstanceId),
		}, nil
	}
	if err != nil {
		return &managerPb.DestroyInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("failed to get instance: %v", err),
		}, nil
	}

	runner := s.getRunnerByURL(instance.RunnerURL)
	if runner == nil {
		return &managerPb.DestroyInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("runner %s not found", instance.RunnerURL),
		}, nil
	}

	runnerReq := &runnerPb.DestroyInstanceRequest{
		InstanceId: req.InstanceId,
	}

	resp, err := runner.Client.DestroyInstance(ctx, runnerReq)
	if err != nil {
		return &managerPb.DestroyInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("failed to destroy instance on runner: %v", err),
		}, nil
	}

	if resp.Status == "success" {
		if err := s.repo.Delete(ctx, req.InstanceId); err != nil {
			log.Printf("Failed to delete instance %s from DB: %v", req.InstanceId, err)
		} else {
			log.Printf("Instance %s destroyed on runner %s", req.InstanceId, runner.URL)
		}
	}

	return &managerPb.DestroyInstanceResponse{
		Status:       resp.Status,
		ErrorMessage: resp.ErrorMessage,
	}, nil
}

func (s *ManagerService) GetInstanceStatus(ctx context.Context, req *managerPb.GetInstanceStatusRequest) (*managerPb.GetInstanceStatusResponse, error) {
	instance, err := s.repo.FindByID(ctx, req.InstanceId)
	if err == domain.ErrInstanceNotFound {
		return &managerPb.GetInstanceStatusResponse{
			State: managerPb.GetInstanceStatusResponse_STATE_DESTROYED,
		}, nil
	}
	if err != nil {
		return &managerPb.GetInstanceStatusResponse{
			State:        managerPb.GetInstanceStatusResponse_STATE_UNSPECIFIED,
			ErrorMessage: fmt.Sprintf("failed to get instance: %v", err),
		}, nil
	}

	runner := s.getRunnerByURL(instance.RunnerURL)
	if runner == nil {
		return &managerPb.GetInstanceStatusResponse{
			State:        managerPb.GetInstanceStatusResponse_STATE_UNSPECIFIED,
			ErrorMessage: fmt.Sprintf("runner %s not found", instance.RunnerURL),
		}, nil
	}

	runnerReq := &runnerPb.GetInstanceStatusRequest{
		InstanceId: req.InstanceId,
	}

	resp, err := runner.Client.GetInstanceStatus(ctx, runnerReq)
	if err != nil {
		return &managerPb.GetInstanceStatusResponse{
			State:        managerPb.GetInstanceStatusResponse_STATE_UNSPECIFIED,
			ErrorMessage: fmt.Sprintf("failed to get instance status from runner: %v", err),
		}, nil
	}

	newState := domain.FromProtoState(managerPb.GetInstanceStatusResponse_State(resp.State))
	if newState != instance.State {
		instance.UpdateState(newState)
		if err := s.repo.Update(ctx, instance); err != nil {
			log.Printf("Failed to update instance %s state: %v", req.InstanceId, err)
		}
	}

	return &managerPb.GetInstanceStatusResponse{
		State:        managerPb.GetInstanceStatusResponse_State(resp.State),
		ErrorMessage: resp.ErrorMessage,
	}, nil
}

func (s *ManagerService) StreamInstanceLogs(req *managerPb.StreamInstanceLogsRequest, stream managerPb.RunnerService_StreamInstanceLogsServer) error {
	ctx := stream.Context()
	instance, err := s.repo.FindByID(ctx, req.InstanceId)
	if err == domain.ErrInstanceNotFound {
		return fmt.Errorf("instance %s not found", req.InstanceId)
	}
	if err != nil {
		return fmt.Errorf("failed to get instance: %v", err)
	}

	runner := s.getRunnerByURL(instance.RunnerURL)
	if runner == nil {
		return fmt.Errorf("runner %s not found", instance.RunnerURL)
	}

	runnerReq := &runnerPb.StreamInstanceLogsRequest{
		InstanceId: req.InstanceId,
	}

	logStream, err := runner.Client.StreamInstanceLogs(ctx, runnerReq)
	if err != nil {
		return fmt.Errorf("failed to stream logs from runner: %v", err)
	}

	for {
		resp, err := logStream.Recv()
		if err != nil {
			return err
		}

		if err := stream.Send(&managerPb.StreamInstanceLogsResponse{
			LogLine: resp.LogLine,
		}); err != nil {
			return err
		}
	}
}

func (s *ManagerService) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx := context.Background()

	instances, err := s.repo.FindAll(ctx)
	if err != nil {
		log.Printf("Failed to get instances: %v", err)
		return
	}

	for _, instance := range instances {
		log.Printf("Cleaning up instance %s", instance.InstanceID)

		if runner := s.getRunnerByURL(instance.RunnerURL); runner != nil {
			req := &runnerPb.DestroyInstanceRequest{
				InstanceId: instance.InstanceID,
			}
			runner.Client.DestroyInstance(ctx, req)
		}

		s.repo.Delete(ctx, instance.InstanceID)
	}

	for _, runner := range s.runners {
		if runner.Connection != nil {
			runner.Connection.Close()
		}
	}
}
