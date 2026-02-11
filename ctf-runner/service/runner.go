package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/netip"
	"os"
	"strconv"

	pb "github.com/kavos113/quickctf/gen/go/api/runner/v1"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
)

type RunnerService struct {
	pb.UnimplementedRunnerServiceServer
	dockerClient *client.Client
	registryURL  string
	minPort      int
	maxPort      int
	usedPorts    []bool
	internalPort int
}

func NewRunnerService(registryURL string) *RunnerService {
	cli, err := client.New(client.FromEnv, client.WithAPIVersionFromEnv())
	if err != nil {
		log.Fatalf("failed to create docker client: %v", err)
	}

	minPort, err := strconv.Atoi(os.Getenv("MIN_OPEN_PORT"))
	if err != nil {
		minPort = 0
	}

	maxPort, err := strconv.Atoi(os.Getenv("MAX_OPEN_PORT"))
	if err != nil {
		maxPort = 0
	}

	internalPort, err := strconv.Atoi(os.Getenv("INTERNAL_CONTAINER_PORT"))
	if err != nil {
		internalPort = 80
	}

	return &RunnerService{
		dockerClient: cli,
		registryURL:  registryURL,
		minPort:      minPort,
		maxPort:      maxPort,
		usedPorts:    make([]bool, maxPort-minPort+1, 0),
		internalPort: internalPort,
	}
}

func (s *RunnerService) selectPort() (int, error) {
	for i, inuse := range s.usedPorts {
		if !inuse {
			s.usedPorts[i] = true
			return i + s.minPort, nil
		}
	}

	return 0, errors.New("no available port")
}

func (s *RunnerService) freePort(p int) {
	if p < s.minPort || p > s.maxPort {
		return
	}

	s.usedPorts[p-s.minPort] = false
}

func (s *RunnerService) StartInstance(ctx context.Context, req *pb.StartInstanceRequest) (*pb.StartInstanceResponse, error) {
	fullImageName := fmt.Sprintf("%s/%s", s.registryURL, req.ImageTag)
	if err := s.pullImage(ctx, fullImageName); err != nil {
		return &pb.StartInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("failed to pull image: %v", err),
		}, nil
	}

	containerPort, _ := network.ParsePort(fmt.Sprintf("%d/tcp", s.internalPort))
	port, err := s.selectPort()
	if err != nil {
		return &pb.StartInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("failed to allocate port: %v", err),
		}, nil
	}
	hostConfig := &container.HostConfig{
		PortBindings: network.PortMap{
			containerPort: []network.PortBinding{
				{
					HostIP:   netip.MustParseAddr("0.0.0.0"),
					HostPort: strconv.Itoa(port),
				},
			},
		},
		AutoRemove: false,
	}

	containerConfig := &container.Config{
		Image: fullImageName,
		ExposedPorts: network.PortSet{
			containerPort: struct{}{},
		},
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

	startOptions := client.ContainerStartOptions{}
	if _, err := s.dockerClient.ContainerStart(ctx, resp.ID, startOptions); err != nil {
		return &pb.StartInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("failed to start container: %v", err),
		}, nil
	}

	inspectOptions := client.ContainerInspectOptions{}
	containerJSON, err := s.dockerClient.ContainerInspect(ctx, resp.ID, inspectOptions)
	if err != nil {
		return &pb.StartInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("failed to inspect container: %v", err),
		}, nil
	}

	var hostPort int32
	if containerJSON.Container.NetworkSettings != nil {
		if bindings, ok := containerJSON.Container.NetworkSettings.Ports[containerPort]; ok && len(bindings) > 0 {
			fmt.Sscanf(bindings[0].HostPort, "%d", &hostPort)
		}
	}

	log.Printf("Started container %s for instance %s on host port %d", resp.ID, req.InstanceId, hostPort)

	return &pb.StartInstanceResponse{
		Status: "success",
		ConnectionInfo: &pb.ConnectionInfo{
			Host: "localhost",
			Port: hostPort,
		},
	}, nil
}

func (s *RunnerService) StopInstance(ctx context.Context, req *pb.StopInstanceRequest) (*pb.StopInstanceResponse, error) {
	timeout := 10
	stopOptions := client.ContainerStopOptions{
		Timeout: &timeout,
	}
	if _, err := s.dockerClient.ContainerStop(ctx, req.InstanceId, stopOptions); err != nil {
		return &pb.StopInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("failed to stop container: %v", err),
		}, nil
	}

	return &pb.StopInstanceResponse{
		Status: "success",
	}, nil
}

func (s *RunnerService) DestroyInstance(ctx context.Context, req *pb.DestroyInstanceRequest) (*pb.DestroyInstanceResponse, error) {
	removeOptions := client.ContainerRemoveOptions{
		Force: true,
	}
	if _, err := s.dockerClient.ContainerRemove(ctx, req.InstanceId, removeOptions); err != nil {
		return &pb.DestroyInstanceResponse{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("failed to remove container: %v", err),
		}, nil
	}

	return &pb.DestroyInstanceResponse{
		Status: "success",
	}, nil
}

func (s *RunnerService) GetInstanceStatus(ctx context.Context, req *pb.GetInstanceStatusRequest) (*pb.GetInstanceStatusResponse, error) {
	inspectOptions := client.ContainerInspectOptions{}
	containerJSON, err := s.dockerClient.ContainerInspect(ctx, req.InstanceId, inspectOptions)
	if err != nil {
		return &pb.GetInstanceStatusResponse{
			State: pb.GetInstanceStatusResponse_STATE_DESTROYED,
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
	ctx := stream.Context()

	logOptions := client.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: false,
	}

	logReader, err := s.dockerClient.ContainerLogs(ctx, req.InstanceId, logOptions)
	if err != nil {
		return fmt.Errorf("failed to get container logs: %v", err)
	}
	defer logReader.Close()

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
	pullOptions := client.ImagePullOptions{}

	reader, err := s.dockerClient.ImagePull(ctx, imageName, pullOptions)
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer reader.Close()

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
