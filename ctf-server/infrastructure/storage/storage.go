package storage

import "context"

type Storage interface {
	Upload(ctx context.Context, bucket, key string, data []byte, contentType string) error
	Download(ctx context.Context, bucket, key string) ([]byte, error)
	Delete(ctx context.Context, bucket, key string) error
	Exists(ctx context.Context, bucket, key string) (bool, error)
	GetPresignedURL(ctx context.Context, bucket, key string) (string, error)
	EnsureBucketExists(ctx context.Context, bucket string) error
}

const (
	BucketBuildLogs   = "build-logs"
	BucketAttachments = "attachments"
)

const (
	BuildLogKeyPrefix   = "logs/"
	AttachmentKeyPrefix = "challenges/"
)
