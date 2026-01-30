package service

import (
	"context"
	"log"
	"net"
	"testing"

	managerPb "github.com/kavos113/quickctf/gen/go/api/manager/v1"
	runnerPb "github.com/kavos113/quickctf/gen/go/api/runner/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var (
	managerLis *bufconn.Listener
	runnerLis  *bufconn.Listener
)

// モックRunnerServiceの実装
type mockRunnerService struct {
	runnerPb.UnimplementedRunnerServiceServer
}

func (m *mockRunnerService) StartInstance(ctx context.Context, req *runnerPb.StartInstanceRequest) (*runnerPb.StartInstanceResponse, error) {
	return &runnerPb.StartInstanceResponse{
		Status: "success",
		ConnectionInfo: &runnerPb.ConnectionInfo{
			Host: "localhost",
			Port: 8080,
		},
	}, nil
}

func (m *mockRunnerService) StopInstance(ctx context.Context, req *runnerPb.StopInstanceRequest) (*runnerPb.StopInstanceResponse, error) {
	return &runnerPb.StopInstanceResponse{
		Status: "success",
	}, nil
}

func (m *mockRunnerService) DestroyInstance(ctx context.Context, req *runnerPb.DestroyInstanceRequest) (*runnerPb.DestroyInstanceResponse, error) {
	return &runnerPb.DestroyInstanceResponse{
		Status: "success",
	}, nil
}

func (m *mockRunnerService) GetInstanceStatus(ctx context.Context, req *runnerPb.GetInstanceStatusRequest) (*runnerPb.GetInstanceStatusResponse, error) {
	return &runnerPb.GetInstanceStatusResponse{
		State: runnerPb.GetInstanceStatusResponse_STATE_RUNNING,
	}, nil
}

func init() {
	// モックRunnerサービスを起動
	runnerLis = bufconn.Listen(bufSize)
	runnerServer := grpc.NewServer()
	runnerPb.RegisterRunnerServiceServer(runnerServer, &mockRunnerService{})
	go func() {
		if err := runnerServer.Serve(runnerLis); err != nil {
			log.Fatalf("Runner server exited with error: %v", err)
		}
	}()

	// Managerサービスを起動
	managerLis = bufconn.Listen(bufSize)
	managerServer := grpc.NewServer()

	// モックリポジトリを使用
	mockRepo := newMockInstanceRepository()

	// テスト用のrunner URLを設定（bufconn経由）
	managerService, err := NewManagerService([]string{"bufnet"}, mockRepo)
	if err != nil {
		log.Fatalf("Failed to create manager service: %v", err)
	}

	// モックRunnerへの接続を上書き
	conn, _ := grpc.Dial("bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return runnerLis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()))

	managerService.runners[0].Connection = conn
	managerService.runners[0].Client = runnerPb.NewRunnerServiceClient(conn)

	managerPb.RegisterRunnerServiceServer(managerServer, managerService)
	go func() {
		if err := managerServer.Serve(managerLis); err != nil {
			log.Fatalf("Manager server exited with error: %v", err)
		}
	}()
}

func managerBufDialer(context.Context, string) (net.Conn, error) {
	return managerLis.Dial()
}

func TestManagerService_StartInstance(t *testing.T) {
	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(managerBufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := managerPb.NewRunnerServiceClient(conn)

	req := &managerPb.StartInstanceRequest{
		ImageTag:   "test:latest",
		InstanceId: "test-instance-1",
		TtlSeconds: 300,
	}

	resp, err := client.StartInstance(ctx, req)
	if err != nil {
		t.Fatalf("StartInstance failed: %v", err)
	}

	if resp.Status != "success" {
		t.Errorf("Expected success status, got %s: %s", resp.Status, resp.ErrorMessage)
	}

	if resp.ConnectionInfo == nil {
		t.Error("Expected connection info, got nil")
	} else {
		t.Logf("Connection: %s:%d", resp.ConnectionInfo.Host, resp.ConnectionInfo.Port)
	}

	// クリーンアップ
	destroyReq := &managerPb.DestroyInstanceRequest{
		InstanceId: "test-instance-1",
	}
	client.DestroyInstance(ctx, destroyReq)
}

func TestManagerService_GetInstanceStatus(t *testing.T) {
	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(managerBufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := managerPb.NewRunnerServiceClient(conn)

	// まずインスタンスを起動
	startReq := &managerPb.StartInstanceRequest{
		ImageTag:   "test:latest",
		InstanceId: "test-instance-2",
		TtlSeconds: 300,
	}
	client.StartInstance(ctx, startReq)

	// ステータスを確認
	statusReq := &managerPb.GetInstanceStatusRequest{
		InstanceId: "test-instance-2",
	}

	resp, err := client.GetInstanceStatus(ctx, statusReq)
	if err != nil {
		t.Fatalf("GetInstanceStatus failed: %v", err)
	}

	if resp.State != managerPb.GetInstanceStatusResponse_STATE_RUNNING {
		t.Errorf("Expected STATE_RUNNING, got %v", resp.State)
	}

	// クリーンアップ
	destroyReq := &managerPb.DestroyInstanceRequest{
		InstanceId: "test-instance-2",
	}
	client.DestroyInstance(ctx, destroyReq)
}

func TestManagerService_DestroyInstance(t *testing.T) {
	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(managerBufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := managerPb.NewRunnerServiceClient(conn)

	// インスタンスを起動
	startReq := &managerPb.StartInstanceRequest{
		ImageTag:   "test:latest",
		InstanceId: "test-instance-3",
		TtlSeconds: 300,
	}
	client.StartInstance(ctx, startReq)

	// 削除
	destroyReq := &managerPb.DestroyInstanceRequest{
		InstanceId: "test-instance-3",
	}

	resp, err := client.DestroyInstance(ctx, destroyReq)
	if err != nil {
		t.Fatalf("DestroyInstance failed: %v", err)
	}

	if resp.Status != "success" {
		t.Errorf("Expected success status, got %s: %s", resp.Status, resp.ErrorMessage)
	}

	// 削除後のステータス確認
	statusReq := &managerPb.GetInstanceStatusRequest{
		InstanceId: "test-instance-3",
	}
	statusResp, _ := client.GetInstanceStatus(ctx, statusReq)

	if statusResp.State != managerPb.GetInstanceStatusResponse_STATE_DESTROYED {
		t.Errorf("Expected STATE_DESTROYED, got %v", statusResp.State)
	}
}
