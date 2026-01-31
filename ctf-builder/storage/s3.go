package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	DefaultBucketName = "build-logs"
	LogFilePrefix     = "logs/"
)

type S3Client struct {
	client     *s3.Client
	bucketName string
}

func NewS3Client() (*S3Client, error) {
	endpoint := os.Getenv("S3_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:9000"
	}

	accessKey := os.Getenv("S3_ACCESS_KEY")
	if accessKey == "" {
		accessKey = "minioadmin"
	}

	secretKey := os.Getenv("S3_SECRET_KEY")
	if secretKey == "" {
		secretKey = "minioadmin"
	}

	bucketName := os.Getenv("S3_BUCKET_NAME")
	if bucketName == "" {
		bucketName = DefaultBucketName
	}

	region := os.Getenv("S3_REGION")
	if region == "" {
		region = "us-east-1"
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	s3Client := &S3Client{
		client:     client,
		bucketName: bucketName,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s3Client.ensureBucketExists(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	return s3Client, nil
}

func (s *S3Client) ensureBucketExists(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucketName),
	})
	if err == nil {
		return nil
	}

	_, err = s.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(s.bucketName),
	})
	if err != nil {
		if strings.Contains(err.Error(), "BucketAlreadyOwnedByYou") ||
			strings.Contains(err.Error(), "BucketAlreadyExists") {
			return nil
		}
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	return nil
}

func (s *S3Client) SaveBuildLog(ctx context.Context, jobID string, logContent string) error {
	key := fmt.Sprintf("%s%s.log", LogFilePrefix, jobID)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader([]byte(logContent)),
		ContentType: aws.String("text/plain; charset=utf-8"),
	})
	if err != nil {
		return fmt.Errorf("failed to save build log: %w", err)
	}

	return nil
}

func (s *S3Client) GetBuildLog(ctx context.Context, jobID string) (string, error) {
	key := fmt.Sprintf("%s%s.log", LogFilePrefix, jobID)

	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get build log: %w", err)
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read build log: %w", err)
	}

	return string(content), nil
}

func (s *S3Client) BuildLogExists(ctx context.Context, jobID string) (bool, error) {
	key := fmt.Sprintf("%s%s.log", LogFilePrefix, jobID)

	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check build log: %w", err)
	}

	return true, nil
}

func (s *S3Client) GetBuildLogURL(ctx context.Context, jobID string) (string, error) {
	key := fmt.Sprintf("%s%s.log", LogFilePrefix, jobID)

	presignClient := s3.NewPresignClient(s.client)

	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(time.Hour))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedReq.URL, nil
}
