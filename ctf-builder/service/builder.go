package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	pb "github.com/kavos113/quickctf/gen/go/api/builder/v1"
	"github.com/moby/moby/client"
)

type BuilderService struct {
	pb.UnimplementedBuilderServiceServer
	dockerClient *client.Client
}

func NewBuilderService() *BuilderService {
	cli, err := client.New(client.FromEnv, client.WithAPIVersionFromEnv())
	if err != nil {
		log.Fatalf("failed to create docker client: %v", err)
	}

	return &BuilderService{
		dockerClient: cli,
	}
}

func (s *BuilderService) BuildImage(req *pb.BuildImageRequest, stream pb.BuilderService_BuildImageServer) error {
	ctx := context.Background()

	// source_tarからイメージをビルド
	buildOptions := client.ImageBuildOptions{
		Tags:       []string{req.ImageTag},
		Dockerfile: "Dockerfile",
		Remove:     true,
		ForceRemove: true,
	}

	sourceTar := bytes.NewReader(req.SourceTar)

	buildResp, err := s.dockerClient.ImageBuild(ctx, sourceTar, buildOptions)
	if err != nil {
		return s.sendError(stream, fmt.Sprintf("failed to build image: %v", err))
	}
	defer buildResp.Body.Close()

	// ビルドログをストリーミング
	decoder := json.NewDecoder(buildResp.Body)
	var buildError string
	var imageID string

	for {
		var message struct {
			Stream string `json:"stream,omitempty"`
			Error  string `json:"error,omitempty"`
			Aux    struct {
				ID string `json:"ID,omitempty"`
			} `json:"aux,omitempty"`
		}

		if err := decoder.Decode(&message); err == io.EOF {
			break
		} else if err != nil {
			return s.sendError(stream, fmt.Sprintf("failed to decode build output: %v", err))
		}

		if message.Error != "" {
			buildError = message.Error
			if err := stream.Send(&pb.BuildImageResponse{
				Response: &pb.BuildImageResponse_LogLine{
					LogLine: message.Error,
				},
			}); err != nil {
				return err
			}
		}

		if message.Stream != "" {
			if err := stream.Send(&pb.BuildImageResponse{
				Response: &pb.BuildImageResponse_LogLine{
					LogLine: message.Stream,
				},
			}); err != nil {
				return err
			}
		}

		if message.Aux.ID != "" {
			imageID = message.Aux.ID
		}
	}

	// ビルド結果を送信
	status := "success"
	if buildError != "" {
		status = "failed"
	}

	result := &pb.BuildResult{
		ImageId:      imageID,
		Status:       status,
		ErrorMessage: buildError,
	}

	return stream.Send(&pb.BuildImageResponse{
		Response: &pb.BuildImageResponse_Result{
			Result: result,
		},
	})
}

func (s *BuilderService) sendError(stream pb.BuilderService_BuildImageServer, errorMsg string) error {
	result := &pb.BuildResult{
		Status:       "failed",
		ErrorMessage: errorMsg,
	}

	if err := stream.Send(&pb.BuildImageResponse{
		Response: &pb.BuildImageResponse_Result{
			Result: result,
		},
	}); err != nil {
		return err
	}

	return fmt.Errorf("%s", errorMsg)
}
