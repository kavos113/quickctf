package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	pb "github.com/kavos113/quickctf/gen/go/api/runner/v1"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
)

type RunnerService struct {
	pb.UnimplementedRunnerServiceServer
	dockerClient *client.Client
	registryURL  string
	instances    map[string]*InstanceInfo
	mu           sync.RWMutex
}

type InstanceInfo struct {
	ContainerID string
	ImageTag    string
	Port        int32
	TTL         time.Duration
	StopTimer   *time.Timer
}

func NewRunnerService(registryURL string) *RunnerService {
	cli, err := client.New(client.FromEnv, client.WithAPIVersionFromEnv())
	if err != nil {
		log.Fatalf("failed to create docker client: %v", err)
	}

	return &RunnerService{
		dockerClient: cli,
		registryURL:  registryURL,
		instances:    make(map[string]*InstanceInfo),
	}
}

func (s *RunnerService) StartInstance(ctx context.Context, req *pb.StartInstanceRequest) (*pb.StartInstanceResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 既に同じインスタンスIDが存在する場合はエラー
	if _, exists := s.instances[req.InstanceId]; exists {
		return &pb.StartInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("instance %s already exists", req.InstanceId),
		}, nil
	}

	// registryからイメージをpull
	fullImageName := fmt.Sprintf("%s/%s", s.registryURL, req.ImageTag)
	if err := s.pullImage(ctx, fullImageName); err != nil {
		return &pb.StartInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("failed to pull image: %v", err),
		}, nil
	}

	// コンテナを作成
	containerConfig := &container.Config{
		Image: fullImageName,
	}

	hostConfig := &container.HostConfig{
		PublishAllPorts: true,
		AutoRemove:      false,
	}

	networkConfig := &network.NetworkingConfig{}

	createOptions := client.ContainerCreateOptions{
		Config:           containerConfig,
		HostConfig:       hostConfig,
		NetworkingConfig: networkConfig,
		Name:             req.InstanceId,
	}

	resp, err := s.dockerClient.ContainerCreate(ctx, createOptions)
	if err != nil {
		return &pb.StartInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("failed to create container: %v", err),
		}, nil
	}

	// コンテナを起動
	startOptions := client.ContainerStartOptions{}
	if _, err := s.dockerClient.ContainerStart(ctx, resp.ID, startOptions); err != nil {
		return &pb.StartInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("failed to start container: %v", err),
		}, nil
	}

	// コンテナ情報を取得してポート番号を取得
	inspectOptions := client.ContainerInspectOptions{}
	containerJSON, err := s.dockerClient.ContainerInspect(ctx, resp.ID, inspectOptions)
	if err != nil {
		return &pb.StartInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("failed to inspect container: %v", err),
		}, nil
	}

	var hostPort int32
	// ポート情報の取得
	if containerJSON.Container.NetworkSettings != nil {
		port, _ := network.ParsePort("80/tcp")
		if bindings, ok := containerJSON.Container.NetworkSettings.Ports[port]; ok && len(bindings) > 0 {
			fmt.Sscanf(bindings[0].HostPort, "%d", &hostPort)
		}
	}

	// インスタンス情報を保存
	ttl := time.Duration(req.TtlSeconds) * time.Second
	instanceInfo := &InstanceInfo{
		ContainerID: resp.ID,
		ImageTag:    req.ImageTag,
		Port:        hostPort,
		TTL:         ttl,
	}

	// TTLが設定されている場合、タイマーを設定
	if req.TtlSeconds > 0 {
		instanceInfo.StopTimer = time.AfterFunc(ttl, func() {
			log.Printf("TTL expired for instance %s, stopping container", req.InstanceId)
			s.stopInstanceInternal(context.Background(), req.InstanceId)
		})
	}

	s.instances[req.InstanceId] = instanceInfo

	return &pb.StartInstanceResponse{
		Status: "success",
		ConnectionInfo: &pb.ConnectionInfo{
			Host: "localhost",
			Port: hostPort,
		},
	}, nil
}

func (s *RunnerService) StopInstance(ctx context.Context, req *pb.StopInstanceRequest) (*pb.StopInstanceResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.stopInstanceInternal(ctx, req.InstanceId), nil
}

func (s *RunnerService) stopInstanceInternal(ctx context.Context, instanceID string) *pb.StopInstanceResponse {
	instanceInfo, exists := s.instances[instanceID]
	if !exists {
		return &pb.StopInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("instance %s not found", instanceID),
		}
	}

	// タイマーをキャンセル
	if instanceInfo.StopTimer != nil {
		instanceInfo.StopTimer.Stop()
	}

	// コンテナを停止
	timeout := 10
	stopOptions := client.ContainerStopOptions{
		Timeout: &timeout,
	}
	if _, err := s.dockerClient.ContainerStop(ctx, instanceInfo.ContainerID, stopOptions); err != nil {
		return &pb.StopInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("failed to stop container: %v", err),
		}
	}

	return &pb.StopInstanceResponse{
		Status: "success",
	}
}

func (s *RunnerService) DestroyInstance(ctx context.Context, req *pb.DestroyInstanceRequest) (*pb.DestroyInstanceResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	instanceInfo, exists := s.instances[req.InstanceId]
	if !exists {
		return &pb.DestroyInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("instance %s not found", req.InstanceId),
		}, nil
	}

	// タイマーをキャンセル
	if instanceInfo.StopTimer != nil {
		instanceInfo.StopTimer.Stop()
	}

	// コンテナを削除
	removeOptions := client.ContainerRemoveOptions{
		Force: true,
	}
	if _, err := s.dockerClient.ContainerRemove(ctx, instanceInfo.ContainerID, removeOptions); err != nil {
		return &pb.DestroyInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("failed to remove container: %v", err),
		}, nil
	}

	// インスタンス情報を削除
	delete(s.instances, req.InstanceId)

	return &pb.DestroyInstanceResponse{
		Status: "success",
	}, nil
}

func (s *RunnerService) GetInstanceStatus(ctx context.Context, req *pb.GetInstanceStatusRequest) (*pb.GetInstanceStatusResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	instanceInfo, exists := s.instances[req.InstanceId]
	if !exists {
		return &pb.GetInstanceStatusResponse{
			State: pb.GetInstanceStatusResponse_STATE_DESTROYED,
		}, nil
	}

	// コンテナの状態を確認
	inspectOptions := client.ContainerInspectOptions{}
	containerJSON, err := s.dockerClient.ContainerInspect(ctx, instanceInfo.ContainerID, inspectOptions)
	if err != nil {
		return &pb.GetInstanceStatusResponse{
			State:        pb.GetInstanceStatusResponse_STATE_UNSPECIFIED,
			ErrorMessage: fmt.Sprintf("failed to inspect container: %v", err),
		}, nil
	}

	var state pb.GetInstanceStatusResponse_State
	if containerJSON.Container.State != nil {
		if containerJSON.Container.State.Running {
			state = pb.GetInstanceStatusResponse_STATE_RUNNING
		} else if containerJSON.Container.State.Status == "exited" {
			state = pb.GetInstanceStatusResponse_STATE_STOPPED
		} else {
			state = pb.GetInstanceStatusResponse_STATE_UNSPECIFIED
		}
	} else {
		state = pb.GetInstanceStatusResponse_STATE_UNSPECIFIED
	}

	return &pb.GetInstanceStatusResponse{
		State: state,
	}, nil
}

func (s *RunnerService) StreamInstanceLogs(req *pb.StreamInstanceLogsRequest, stream pb.RunnerService_StreamInstanceLogsServer) error {
	s.mu.RLock()
	instanceInfo, exists := s.instances[req.InstanceId]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("instance %s not found", req.InstanceId)
	}

	ctx := stream.Context()

	// コンテナのログをストリーミング
	logOptions := client.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: false,
	}

	logReader, err := s.dockerClient.ContainerLogs(ctx, instanceInfo.ContainerID, logOptions)
	if err != nil {
		return fmt.Errorf("failed to get container logs: %v", err)
	}
	defer logReader.Close()

	// ログを読み取って送信
	buf := make([]byte, 8192)
	for {
		n, err := logReader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read logs: %v", err)
		}

		if n > 0 {
			// Docker APIはログの先頭8バイトにヘッダーを含むため、それをスキップ
			logLine := string(buf[8:n])
			if err := stream.Send(&pb.StreamInstanceLogsResponse{
				LogLine: logLine,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *RunnerService) pullImage(ctx context.Context, imageName string) error {
	// 認証情報を作成（匿名アクセスの場合）
	authConfig := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}
	authJSON, err := json.Marshal(authConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal auth config: %w", err)
	}
	encodedAuth := base64.URLEncoding.EncodeToString(authJSON)

	pullOptions := client.ImagePullOptions{
		RegistryAuth: encodedAuth,
	}

	reader, err := s.dockerClient.ImagePull(ctx, imageName, pullOptions)
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer reader.Close()

	// pullの完了を待つ
	decoder := json.NewDecoder(reader)
	for {
		var message struct {
			Status string `json:"status,omitempty"`
			Error  string `json:"error,omitempty"`
		}

		if err := decoder.Decode(&message); err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("failed to decode pull output: %w", err)
		}

		if message.Error != "" {
			return fmt.Errorf("pull error: %s", message.Error)
		}

		log.Printf("Pull: %s", message.Status)
	}

	return nil
}

func (s *RunnerService) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx := context.Background()
	for instanceID, instanceInfo := range s.instances {
		log.Printf("Cleaning up instance %s", instanceID)
		if instanceInfo.StopTimer != nil {
			instanceInfo.StopTimer.Stop()
		}
		
		removeOptions := client.ContainerRemoveOptions{
			Force: true,
		}
		s.dockerClient.ContainerRemove(ctx, instanceInfo.ContainerID, removeOptions)
	}
}
