package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	pb "github.com/kavos113/quickctf/gen/go/api/builder/v1"
	"github.com/moby/moby/client"
)

type BuilderService struct {
	pb.UnimplementedBuilderServiceServer
	dockerClient *client.Client
	registryURL  string
}

func NewBuilderService() *BuilderService {
	cli, err := client.New(client.FromEnv, client.WithAPIVersionFromEnv())
	if err != nil {
		log.Fatalf("failed to create docker client: %v", err)
	}

	registryURL := os.Getenv("CTF_REGISTRY_URL")
	if registryURL == "" {
		registryURL = "localhost:5000"
	}

	return &BuilderService{
		dockerClient: cli,
		registryURL:  registryURL,
	}
}

func (s *BuilderService) BuildImage(req *pb.BuildImageRequest, stream pb.BuilderService_BuildImageServer) error {
	ctx := context.Background()

	// source_tarからイメージをビルド
	buildOptions := client.ImageBuildOptions{
		Tags:        []string{req.ImageTag},
		Dockerfile:  "Dockerfile",
		Remove:      true,
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

	// ビルドエラーがある場合は結果を送信して終了
	if buildError != "" {
		result := &pb.BuildResult{
			ImageId:      imageID,
			Status:       "failed",
			ErrorMessage: buildError,
		}
		return stream.Send(&pb.BuildImageResponse{
			Response: &pb.BuildImageResponse_Result{
				Result: result,
			},
		})
	}

	// イメージをregistryにpush
	if err := stream.Send(&pb.BuildImageResponse{
		Response: &pb.BuildImageResponse_LogLine{
			LogLine: fmt.Sprintf("Pushing image to registry %s...\n", s.registryURL),
		},
	}); err != nil {
		return err
	}

	if err := s.pushImage(ctx, req.ImageTag, stream); err != nil {
		return s.sendError(stream, fmt.Sprintf("failed to push image: %v", err))
	}

	// 成功結果を送信
	result := &pb.BuildResult{
		ImageId:      imageID,
		Status:       "success",
		ErrorMessage: "",
	}

	return stream.Send(&pb.BuildImageResponse{
		Response: &pb.BuildImageResponse_Result{
			Result: result,
		},
	})
}

func (s *BuilderService) pushImage(ctx context.Context, imageTag string, stream pb.BuilderService_BuildImageServer) error {
	// registryのURLを含むタグを作成
	registryTag := fmt.Sprintf("%s/%s", s.registryURL, imageTag)

	// イメージにregistryタグを付ける
	tagOptions := client.ImageTagOptions{
		Source: imageTag,
		Target: registryTag,
	}
	if _, err := s.dockerClient.ImageTag(ctx, tagOptions); err != nil {
		return fmt.Errorf("failed to tag image: %w", err)
	}

	// イメージをpush
	pushOptions := client.ImagePushOptions{}

	pushResp, err := s.dockerClient.ImagePush(ctx, registryTag, pushOptions)
	if err != nil {
		return fmt.Errorf("failed to push image: %w", err)
	}
	defer pushResp.Close()

	// pushログをストリーミング
	decoder := json.NewDecoder(pushResp)
	for {
		var message struct {
			Status   string `json:"status,omitempty"`
			Progress string `json:"progress,omitempty"`
			Error    string `json:"error,omitempty"`
		}

		if err := decoder.Decode(&message); err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("failed to decode push output: %w", err)
		}

		if message.Error != "" {
			return fmt.Errorf("push error: %s", message.Error)
		}

		if message.Status != "" {
			logLine := message.Status
			if message.Progress != "" {
				logLine = fmt.Sprintf("%s %s", message.Status, message.Progress)
			}
			if err := stream.Send(&pb.BuildImageResponse{
				Response: &pb.BuildImageResponse_LogLine{
					LogLine: logLine + "\n",
				},
			}); err != nil {
				return err
			}
		}
	}

	return nil
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
