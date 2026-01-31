package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/kavos113/quickctf/ctf-builder/queue"
	"github.com/moby/moby/client"
)

type BuildWorker struct {
	dockerClient *client.Client
	redisClient  *queue.RedisClient
	registryURL  string
}

func NewBuildWorker(redisClient *queue.RedisClient) *BuildWorker {
	cli, err := client.New(client.FromEnv, client.WithAPIVersionFromEnv())
	if err != nil {
		log.Fatalf("failed to create docker client: %v", err)
	}

	registryURL := os.Getenv("CTF_REGISTRY_URL")
	if registryURL == "" {
		registryURL = "localhost:5000"
	}

	return &BuildWorker{
		dockerClient: cli,
		redisClient:  redisClient,
		registryURL:  registryURL,
	}
}

func (w *BuildWorker) Start(ctx context.Context) {
	log.Println("Build worker started, waiting for jobs...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Build worker stopping...")
			return
		default:
			job, err := w.redisClient.DequeueJob(ctx, 5*time.Second)
			if err != nil {
				log.Printf("Error dequeuing job: %v", err)
				continue
			}
			if job == nil {
				// No job available, continue waiting
				continue
			}

			log.Printf("Processing job %s for image %s", job.JobID, job.ImageTag)
			w.processJob(ctx, job)
		}
	}
}

func (w *BuildWorker) processJob(ctx context.Context, job *queue.BuildJob) {
	result := &queue.BuildResult{
		JobID:  job.JobID,
		Status: queue.BuildStatusBuilding,
	}
	w.redisClient.SetBuildResult(ctx, result)

	imageID, err := w.buildImage(ctx, job)
	if err != nil {
		log.Printf("Build failed for job %s: %v", job.JobID, err)
		result = &queue.BuildResult{
			JobID:        job.JobID,
			Status:       queue.BuildStatusFailed,
			ErrorMessage: err.Error(),
			CompletedAt:  time.Now(),
		}
		w.redisClient.SetBuildResult(ctx, result)
		w.redisClient.PublishLog(ctx, job.JobID, fmt.Sprintf("BUILD_COMPLETE:failed:%s", err.Error()))
		return
	}

	w.redisClient.PublishLog(ctx, job.JobID, fmt.Sprintf("Pushing image to registry %s...\n", w.registryURL))
	if err := w.pushImage(ctx, job); err != nil {
		log.Printf("Push failed for job %s: %v", job.JobID, err)
		result = &queue.BuildResult{
			JobID:        job.JobID,
			ImageID:      imageID,
			Status:       queue.BuildStatusFailed,
			ErrorMessage: fmt.Sprintf("push failed: %v", err),
			CompletedAt:  time.Now(),
		}
		w.redisClient.SetBuildResult(ctx, result)
		w.redisClient.PublishLog(ctx, job.JobID, fmt.Sprintf("BUILD_COMPLETE:failed:push failed: %v", err))
		return
	}

	log.Printf("Build succeeded for job %s, image ID: %s", job.JobID, imageID)
	result = &queue.BuildResult{
		JobID:       job.JobID,
		ImageID:     imageID,
		Status:      queue.BuildStatusSuccess,
		CompletedAt: time.Now(),
	}
	w.redisClient.SetBuildResult(ctx, result)
	w.redisClient.PublishLog(ctx, job.JobID, fmt.Sprintf("BUILD_COMPLETE:success:%s", imageID))
}

func (w *BuildWorker) buildImage(ctx context.Context, job *queue.BuildJob) (string, error) {
	buildOptions := client.ImageBuildOptions{
		Tags:        []string{job.ImageTag},
		Dockerfile:  "Dockerfile",
		Remove:      true,
		ForceRemove: true,
	}

	sourceTar := bytes.NewReader(job.SourceTar)

	buildResp, err := w.dockerClient.ImageBuild(ctx, sourceTar, buildOptions)
	if err != nil {
		return "", fmt.Errorf("failed to build image: %w", err)
	}
	defer buildResp.Body.Close()

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
			return "", fmt.Errorf("failed to decode build output: %w", err)
		}

		if message.Error != "" {
			buildError = message.Error
			w.redisClient.PublishLog(ctx, job.JobID, message.Error)
		}

		if message.Stream != "" {
			w.redisClient.PublishLog(ctx, job.JobID, message.Stream)
		}

		if message.Aux.ID != "" {
			imageID = message.Aux.ID
		}
	}

	if buildError != "" {
		return imageID, fmt.Errorf("build error: %s", buildError)
	}

	return imageID, nil
}

func (w *BuildWorker) pushImage(ctx context.Context, job *queue.BuildJob) error {
	registryTag := fmt.Sprintf("%s/%s", w.registryURL, job.ImageTag)

	tagOptions := client.ImageTagOptions{
		Source: job.ImageTag,
		Target: registryTag,
	}
	if _, err := w.dockerClient.ImageTag(ctx, tagOptions); err != nil {
		return fmt.Errorf("failed to tag image: %w", err)
	}

	pushOptions := client.ImagePushOptions{}

	pushResp, err := w.dockerClient.ImagePush(ctx, registryTag, pushOptions)
	if err != nil {
		return fmt.Errorf("failed to push image: %w", err)
	}
	defer pushResp.Close()

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
			w.redisClient.PublishLog(ctx, job.JobID, logLine+"\n")
		}
	}

	return nil
}
