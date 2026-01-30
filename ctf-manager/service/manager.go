package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	managerPb "github.com/kavos113/quickctf/gen/go/api/manager/v1"
	runnerPb "github.com/kavos113/quickctf/gen/go/api/runner/v1"
)

type ManagerService struct {
	managerPb.UnimplementedRunnerServiceServer
	runners    []*RunnerClient
	instances  map[string]*InstanceInfo
	mu         sync.RWMutex
	nextRunner int
}

type RunnerClient struct {
	URL        string
	Client     runnerPb.RunnerServiceClient
	Connection *grpc.ClientConn
	Active     bool
}

type InstanceInfo struct {
	InstanceID     string
	ImageTag       string
	RunnerURL      string
	ConnectionInfo *managerPb.ConnectionInfo
	State          managerPb.GetInstanceStatusResponse_State
	TTL            time.Duration
	CreatedAt      time.Time
}

func NewManagerService(runnerURLs []string) (*ManagerService, error) {
	runners := make([]*RunnerClient, 0, len(runnerURLs))

	for _, url := range runnerURLs {
		conn, err := grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
		runners:   runners,
		instances: make(map[string]*InstanceInfo),
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
	s.mu.Lock()

	// 既に同じインスタンスIDが存在する場合はエラー
	if _, exists := s.instances[req.InstanceId]; exists {
		s.mu.Unlock()
		return &managerPb.StartInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("instance %s already exists", req.InstanceId),
		}, nil
	}
	s.mu.Unlock()

	// 利用可能なrunnerを選択
	runner := s.selectRunner()
	if runner == nil {
		return &managerPb.StartInstanceResponse{
			Status:       "failed",
			ErrorMessage: "no available runners",
		}, nil
	}

	// runnerにリクエストを転送
	runnerReq := &runnerPb.StartInstanceRequest{
		ImageTag:   req.ImageTag,
		InstanceId: req.InstanceId,
		TtlSeconds: req.TtlSeconds,
	}

	resp, err := runner.Client.StartInstance(ctx, runnerReq)
	if err != nil {
		return &managerPb.StartInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("failed to start instance on runner: %v", err),
		}, nil
	}

	// 成功した場合、インスタンス情報を保存
	if resp.Status == "success" {
		s.mu.Lock()
		s.instances[req.InstanceId] = &InstanceInfo{
			InstanceID: req.InstanceId,
			ImageTag:   req.ImageTag,
			RunnerURL:  runner.URL,
			ConnectionInfo: &managerPb.ConnectionInfo{
				Host: resp.ConnectionInfo.Host,
				Port: resp.ConnectionInfo.Port,
			},
			State:     managerPb.GetInstanceStatusResponse_STATE_RUNNING,
			TTL:       time.Duration(req.TtlSeconds) * time.Second,
			CreatedAt: time.Now(),
		}
		s.mu.Unlock()

		log.Printf("Instance %s started on runner %s", req.InstanceId, runner.URL)
	}

	// ConnectionInfoを変換
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
	s.mu.RLock()
	instance, exists := s.instances[req.InstanceId]
	s.mu.RUnlock()

	if !exists {
		return &managerPb.StopInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("instance %s not found", req.InstanceId),
		}, nil
	}

	runner := s.getRunnerByURL(instance.RunnerURL)
	if runner == nil {
		return &managerPb.StopInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("runner %s not found", instance.RunnerURL),
		}, nil
	}

	// runnerにリクエストを転送
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

	// 状態を更新
	if resp.Status == "success" {
		s.mu.Lock()
		if inst, ok := s.instances[req.InstanceId]; ok {
			inst.State = managerPb.GetInstanceStatusResponse_STATE_STOPPED
		}
		s.mu.Unlock()

		log.Printf("Instance %s stopped on runner %s", req.InstanceId, runner.URL)
	}

	return &managerPb.StopInstanceResponse{
		Status:       resp.Status,
		ErrorMessage: resp.ErrorMessage,
	}, nil
}

func (s *ManagerService) DestroyInstance(ctx context.Context, req *managerPb.DestroyInstanceRequest) (*managerPb.DestroyInstanceResponse, error) {
	s.mu.RLock()
	instance, exists := s.instances[req.InstanceId]
	s.mu.RUnlock()

	if !exists {
		return &managerPb.DestroyInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("instance %s not found", req.InstanceId),
		}, nil
	}

	runner := s.getRunnerByURL(instance.RunnerURL)
	if runner == nil {
		return &managerPb.DestroyInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("runner %s not found", instance.RunnerURL),
		}, nil
	}

	// runnerにリクエストを転送
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

	// 成功した場合、インスタンス情報を削除
	if resp.Status == "success" {
		s.mu.Lock()
		delete(s.instances, req.InstanceId)
		s.mu.Unlock()

		log.Printf("Instance %s destroyed on runner %s", req.InstanceId, runner.URL)
	}

	return &managerPb.DestroyInstanceResponse{
		Status:       resp.Status,
		ErrorMessage: resp.ErrorMessage,
	}, nil
}

func (s *ManagerService) GetInstanceStatus(ctx context.Context, req *managerPb.GetInstanceStatusRequest) (*managerPb.GetInstanceStatusResponse, error) {
	s.mu.RLock()
	instance, exists := s.instances[req.InstanceId]
	s.mu.RUnlock()

	if !exists {
		return &managerPb.GetInstanceStatusResponse{
			State: managerPb.GetInstanceStatusResponse_STATE_DESTROYED,
		}, nil
	}

	runner := s.getRunnerByURL(instance.RunnerURL)
	if runner == nil {
		return &managerPb.GetInstanceStatusResponse{
			State:        managerPb.GetInstanceStatusResponse_STATE_UNSPECIFIED,
			ErrorMessage: fmt.Sprintf("runner %s not found", instance.RunnerURL),
		}, nil
	}

	// runnerにリクエストを転送
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

	// 状態を更新
	s.mu.Lock()
	if inst, ok := s.instances[req.InstanceId]; ok {
		inst.State = managerPb.GetInstanceStatusResponse_State(resp.State)
	}
	s.mu.Unlock()

	return &managerPb.GetInstanceStatusResponse{
		State:        managerPb.GetInstanceStatusResponse_State(resp.State),
		ErrorMessage: resp.ErrorMessage,
	}, nil
}

func (s *ManagerService) StreamInstanceLogs(req *managerPb.StreamInstanceLogsRequest, stream managerPb.RunnerService_StreamInstanceLogsServer) error {
	s.mu.RLock()
	instance, exists := s.instances[req.InstanceId]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("instance %s not found", req.InstanceId)
	}

	runner := s.getRunnerByURL(instance.RunnerURL)
	if runner == nil {
		return fmt.Errorf("runner %s not found", instance.RunnerURL)
	}

	// runnerにリクエストを転送
	ctx := stream.Context()
	runnerReq := &runnerPb.StreamInstanceLogsRequest{
		InstanceId: req.InstanceId,
	}

	logStream, err := runner.Client.StreamInstanceLogs(ctx, runnerReq)
	if err != nil {
		return fmt.Errorf("failed to stream logs from runner: %v", err)
	}

	// runnerからのログをクライアントに転送
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

	// 全てのインスタンスを削除
	for instanceID, instance := range s.instances {
		log.Printf("Cleaning up instance %s", instanceID)

		if runner := s.getRunnerByURL(instance.RunnerURL); runner != nil {
			req := &runnerPb.DestroyInstanceRequest{
				InstanceId: instanceID,
			}
			runner.Client.DestroyInstance(ctx, req)
		}
	}

	// runner接続を閉じる
	for _, runner := range s.runners {
		if runner.Connection != nil {
			runner.Connection.Close()
		}
	}
}
