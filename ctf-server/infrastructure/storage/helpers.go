package storage

import (
	"context"
	"fmt"
)

type AttachmentStorage struct {
	storage Storage
}

func NewAttachmentStorage(storage Storage) *AttachmentStorage {
	return &AttachmentStorage{storage: storage}
}

func (a *AttachmentStorage) EnsureBucket(ctx context.Context) error {
	return a.storage.EnsureBucketExists(ctx, BucketAttachments)
}

func (a *AttachmentStorage) Upload(ctx context.Context, challengeID, filename string, data []byte) (string, error) {
	key := fmt.Sprintf("%s%s/%s", AttachmentKeyPrefix, challengeID, filename)

	err := a.storage.Upload(ctx, BucketAttachments, key, data, "")
	if err != nil {
		return "", fmt.Errorf("failed to upload attachment: %w", err)
	}

	return key, nil
}

func (a *AttachmentStorage) Delete(ctx context.Context, key string) error {
	return a.storage.Delete(ctx, BucketAttachments, key)
}

func (a *AttachmentStorage) GetPresignedURL(ctx context.Context, key string) (string, error) {
	return a.storage.GetPresignedURL(ctx, BucketAttachments, key)
}

type BuildLogStorage struct {
	storage Storage
}

func NewBuildLogStorage(storage Storage) *BuildLogStorage {
	return &BuildLogStorage{storage: storage}
}

func (b *BuildLogStorage) EnsureBucket(ctx context.Context) error {
	return b.storage.EnsureBucketExists(ctx, BucketBuildLogs)
}

func (b *BuildLogStorage) GetLog(ctx context.Context, jobID string) (string, error) {
	key := fmt.Sprintf("%s%s.log", BuildLogKeyPrefix, jobID)
	data, err := b.storage.Download(ctx, BucketBuildLogs, key)
	if err != nil {
		return "", fmt.Errorf("failed to get build log: %w", err)
	}
	return string(data), nil
}
