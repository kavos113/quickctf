package queue

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	BuildQueueKey    = "build:queue"
	BuildResultKey   = "build:result:"  // + job_id
	BuildLogChannel  = "build:logs:"    // + job_id
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient() (*RedisClient, error) {
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

	return &RedisClient{client: client}, nil
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

func (r *RedisClient) EnqueueJob(ctx context.Context, job *BuildJob) error {
	data, err := job.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	if err := r.client.RPush(ctx, BuildQueueKey, data).Err(); err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	result := &BuildResult{
		JobID:  job.JobID,
		Status: BuildStatusPending,
	}
	if err := r.SetBuildResult(ctx, result); err != nil {
		return fmt.Errorf("failed to set initial status: %w", err)
	}

	return nil
}

func (r *RedisClient) DequeueJob(ctx context.Context, timeout time.Duration) (*BuildJob, error) {
	result, err := r.client.BLPop(ctx, timeout, BuildQueueKey).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("unexpected result format")
	}

	return ParseBuildJob([]byte(result[1]))
}

func (r *RedisClient) SetBuildResult(ctx context.Context, result *BuildResult) error {
	data, err := result.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	key := BuildResultKey + result.JobID
	if err := r.client.Set(ctx, key, data, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to set result: %w", err)
	}

	return nil
}

func (r *RedisClient) GetBuildResult(ctx context.Context, jobID string) (*BuildResult, error) {
	key := BuildResultKey + jobID
	data, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get result: %w", err)
	}

	return ParseBuildResult([]byte(data))
}

func (r *RedisClient) PublishLog(ctx context.Context, jobID, logLine string) error {
	channel := BuildLogChannel + jobID
	if err := r.client.Publish(ctx, channel, logLine).Err(); err != nil {
		return fmt.Errorf("failed to publish log: %w", err)
	}
	return nil
}

func (r *RedisClient) SubscribeLogs(ctx context.Context, jobID string) *redis.PubSub {
	channel := BuildLogChannel + jobID
	return r.client.Subscribe(ctx, channel)
}
