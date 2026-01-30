package service

import (
	"context"
	"log"
	"net"
	"testing"

	pb "github.com/kavos113/quickctf/gen/go/api/runner/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterRunnerServiceServer(s, NewRunnerService("localhost:5000"))
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestRunnerService_StartInstance(t *testing.T) {
	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewRunnerServiceClient(conn)

	req := &pb.StartInstanceRequest{
		ImageTag:   "test:latest",
		InstanceId: "test-instance-1",
		TtlSeconds: 300,
	}

	resp, err := client.StartInstance(ctx, req)
	if err != nil {
		t.Fatalf("StartInstance failed: %v", err)
	}

	t.Logf("Status: %s", resp.Status)
	if resp.ErrorMessage != "" {
		t.Logf("Error: %s", resp.ErrorMessage)
	}

	if resp.Status == "success" {
		t.Log("Unexpected success - Docker daemon might be running")
		destroyReq := &pb.DestroyInstanceRequest{
			InstanceId: "test-instance-1",
		}
		client.DestroyInstance(ctx, destroyReq)
	}
}

func TestRunnerService_GetInstanceStatus(t *testing.T) {
	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewRunnerServiceClient(conn)

	req := &pb.GetInstanceStatusRequest{
		InstanceId: "non-existent-instance",
	}

	resp, err := client.GetInstanceStatus(ctx, req)
	if err != nil {
		t.Fatalf("GetInstanceStatus failed: %v", err)
	}

	if resp.State != pb.GetInstanceStatusResponse_STATE_DESTROYED {
		t.Errorf("Expected STATE_DESTROYED, got %v", resp.State)
	}
}

func TestRunnerService_StopInstance(t *testing.T) {
	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewRunnerServiceClient(conn)

	req := &pb.StopInstanceRequest{
		InstanceId: "non-existent-instance",
	}

	resp, err := client.StopInstance(ctx, req)
	if err != nil {
		t.Fatalf("StopInstance failed: %v", err)
	}

	if resp.Status != "failed" {
		t.Errorf("Expected failed status, got %s", resp.Status)
	}
}
