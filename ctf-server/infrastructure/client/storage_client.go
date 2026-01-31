package client

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	AttachmentBucketName   = "attachments"
	AttachmentKeyPrefix    = "challenges/"
	PresignedURLExpiration = 1 * time.Hour
)

type StorageClient struct {
	s3Client   *s3.Client
	bucketName string
	s3Endpoint string
}

func NewStorageClient() (*StorageClient, error) {
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

	bucketName := os.Getenv("S3_ATTACHMENT_BUCKET")
	if bucketName == "" {
		bucketName = AttachmentBucketName
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

	storage := &StorageClient{
		s3Client:   client,
		bucketName: bucketName,
		s3Endpoint: endpoint,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := storage.ensureBucketExists(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	return storage, nil
}

func (s *StorageClient) ensureBucketExists(ctx context.Context) error {
	_, err := s.s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucketName),
	})
	if err == nil {
		return nil
	}

	_, err = s.s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
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

func (s *StorageClient) UploadAttachment(ctx context.Context, challengeID, filename string, data []byte) (string, error) {
	key := fmt.Sprintf("%s%s/%s", AttachmentKeyPrefix, challengeID, filename)

	_, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:             aws.String(s.bucketName),
		Key:                aws.String(key),
		Body:               bytes.NewReader(data),
		ContentDisposition: aws.String(fmt.Sprintf("attachment; filename=\"%s\"", filename)),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload attachment: %w", err)
	}

	return key, nil
}

func (s *StorageClient) DeleteAttachment(ctx context.Context, s3Key string) error {
	_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete attachment: %w", err)
	}

	return nil
}

func (s *StorageClient) GetPresignedURL(ctx context.Context, s3Key string) (string, error) {
	presignClient := s3.NewPresignClient(s.s3Client)

	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(s3Key),
	}, s3.WithPresignExpires(PresignedURLExpiration))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedReq.URL, nil
}
