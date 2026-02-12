package client

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/kavos113/quickctf/ctf-server/domain"
	pb "github.com/kavos113/quickctf/gen/go/api/manager/v1"
)

type ManagerClient struct {
	client pb.RunnerServiceClient
	conn   *grpc.ClientConn
}

func NewManagerClient() (*ManagerClient, error) {
	managerAddr := os.Getenv("MANAGER_ADDRESS")
	if managerAddr == "" {
		managerAddr = "localhost:50052"
	}

	conn, err := grpc.NewClient(managerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to manager service: %w", err)
	}

	client := pb.NewRunnerServiceClient(conn)

	return &ManagerClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *ManagerClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *ManagerClient) StartInstance(ctx context.Context, imageTag string, ttlSeconds int64) (string, *pb.ConnectionInfo, error) {
	resp, err := c.client.StartInstance(ctx, &pb.StartInstanceRequest{
		ImageTag:   imageTag,
		TtlSeconds: ttlSeconds,
	})
	if err != nil {
		return "", nil, fmt.Errorf("failed to start instance: %w", err)
	}

	if resp.Status != "success" && resp.Status != "running" {
		return "", nil, fmt.Errorf("start instance failed: %s", resp.ErrorMessage)
	}

	return resp.InstanceId, resp.ConnectionInfo, nil
}

func (c *ManagerClient) StopInstance(ctx context.Context, instanceID string) error {
	resp, err := c.client.StopInstance(ctx, &pb.StopInstanceRequest{
		InstanceId: instanceID,
	})
	if err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}

	if resp.Status != "success" && resp.Status != "stopped" {
		return fmt.Errorf("stop instance failed: %s", resp.ErrorMessage)
	}

	return nil
}

func (c *ManagerClient) DestroyInstance(ctx context.Context, instanceID string) error {
	resp, err := c.client.DestroyInstance(ctx, &pb.DestroyInstanceRequest{
		InstanceId: instanceID,
	})
	if err != nil {
		return fmt.Errorf("failed to destroy instance: %w", err)
	}

	if resp.Status != "success" && resp.Status != "destroyed" {
		return fmt.Errorf("destroy instance failed: %s", resp.ErrorMessage)
	}

	return nil
}

func (c *ManagerClient) GetInstanceStatus(ctx context.Context, instanceID string) (domain.InstanceStatus, error) {
	resp, err := c.client.GetInstanceStatus(ctx, &pb.GetInstanceStatusRequest{
		InstanceId: instanceID,
	})
	if err != nil {
		return domain.InstanceStatusUnknown, fmt.Errorf("failed to get instance status: %w", err)
	}

	if resp.ErrorMessage != "" {
		return domain.InstanceStatusUnknown, fmt.Errorf("get instance status failed: %s", resp.ErrorMessage)
	}

	switch resp.State {
	case pb.GetInstanceStatusResponse_STATE_RUNNING:
		return domain.InstanceStatusRunning, nil
	case pb.GetInstanceStatusResponse_STATE_STOPPED:
		return domain.InstanceStatusStopped, nil
	case pb.GetInstanceStatusResponse_STATE_DESTROYED:
		return domain.InstanceStatusDestroyed, nil
	default:
		return domain.InstanceStatusUnknown, nil
	}
}
