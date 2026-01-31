package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	BuildQueueKey   = "build:queue"
	BuildResultKey  = "build:result:"
	BuildLogChannel = "build:logs:"
)

type BuildJob struct {
	JobID       string    `json:"job_id"`
	ImageTag    string    `json:"image_tag"`
	SourceTar   []byte    `json:"source_tar"`
	CreatedAt   time.Time `json:"created_at"`
	ChallengeID string    `json:"challenge_id"`
}

type BuildResult struct {
	JobID        string    `json:"job_id"`
	ImageID      string    `json:"image_id"`
	Status       string    `json:"status"`
	ErrorMessage string    `json:"error_message,omitempty"`
	CompletedAt  time.Time `json:"completed_at,omitempty"`
}

const (
	BuildStatusPending  = "pending"
	BuildStatusBuilding = "building"
	BuildStatusSuccess  = "success"
	BuildStatusFailed   = "failed"
)

type BuilderClient struct {
	redisClient *redis.Client
}

func NewBuilderClient() (*BuilderClient, error) {
	redisAddr := os.Getenv("REDIS_ADDRESS")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &BuilderClient{redisClient: client}, nil
}

func (c *BuilderClient) Close() error {
	return c.redisClient.Close()
}

func (c *BuilderClient) EnqueueBuild(ctx context.Context, imageTag string, sourceTar []byte, challengeID string) (string, error) {
	jobID := uuid.New().String()

	job := &BuildJob{
		JobID:       jobID,
		ImageTag:    imageTag,
		SourceTar:   sourceTar,
		CreatedAt:   time.Now(),
		ChallengeID: challengeID,
	}

	data, err := json.Marshal(job)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job: %w", err)
	}

	if err := c.redisClient.RPush(ctx, BuildQueueKey, data).Err(); err != nil {
		return "", fmt.Errorf("failed to enqueue job: %w", err)
	}

	result := &BuildResult{
		JobID:  jobID,
		Status: BuildStatusPending,
	}
	resultData, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	key := BuildResultKey + jobID
	if err := c.redisClient.Set(ctx, key, resultData, 24*time.Hour).Err(); err != nil {
		return "", fmt.Errorf("failed to set initial status: %w", err)
	}

	return jobID, nil
}

func (c *BuilderClient) GetBuildResult(ctx context.Context, jobID string) (*BuildResult, error) {
	key := BuildResultKey + jobID
	data, err := c.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get result: %w", err)
	}

	var result BuildResult
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}

func (c *BuilderClient) SubscribeBuildLogs(ctx context.Context, jobID string, callback func(logLine string)) error {
	channel := BuildLogChannel + jobID
	pubsub := c.redisClient.Subscribe(ctx, channel)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-ch:
			if !ok {
				return nil
			}
			callback(msg.Payload)

			if strings.HasPrefix(msg.Payload, "BUILD_COMPLETE:") {
				return nil
			}
		}
	}
}

func (c *BuilderClient) BuildImage(ctx context.Context, imageTag string, sourceTar []byte) (string, error) {
	jobID, err := c.EnqueueBuild(ctx, imageTag, sourceTar, "")
	if err != nil {
		return "", fmt.Errorf("failed to enqueue build: %w", err)
	}

	log.Printf("Build job %s enqueued for image %s", jobID, imageTag)

	logsDone := make(chan struct{})
	go func() {
		defer close(logsDone)
		err := c.SubscribeBuildLogs(ctx, jobID, func(logLine string) {
			log.Printf("[Builder] %s", logLine)
		})
		if err != nil && err != context.Canceled {
			log.Printf("Error subscribing to build logs: %v", err)
		}
	}()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-logsDone:
			result, err := c.GetBuildResult(ctx, jobID)
			if err != nil {
				return "", fmt.Errorf("failed to get build result: %w", err)
			}
			if result == nil {
				return "", fmt.Errorf("build result not found")
			}
			if result.Status == BuildStatusSuccess {
				return result.ImageID, nil
			}
			return "", fmt.Errorf("build failed: %s", result.ErrorMessage)
		case <-ticker.C:
			result, err := c.GetBuildResult(ctx, jobID)
			if err != nil {
				log.Printf("Error getting build result: %v", err)
				continue
			}
			if result == nil {
				continue
			}
			if result.Status == BuildStatusSuccess {
				return result.ImageID, nil
			}
			if result.Status == BuildStatusFailed {
				return "", fmt.Errorf("build failed: %s", result.ErrorMessage)
			}
		}
	}
}
