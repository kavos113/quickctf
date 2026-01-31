package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/kavos113/quickctf/ctf-registry/storage"
	"github.com/opencontainers/go-digest"
)

type Storage struct {
	client *s3.Client
	bucket string
}

func NewStorage() *Storage {
	bucket := os.Getenv("S3_BUCKET")
	if bucket == "" {
		bucket = "ctf-registry"
	}

	endpoint := os.Getenv("S3_ENDPOINT")
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	var cfg aws.Config
	var err error

	ctx := context.Background()

	if endpoint != "" {
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		)
		if err != nil {
			log.Fatalf("failed to load AWS config: %v", err)
		}
	} else {
		cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(region))
		if err != nil {
			log.Fatalf("failed to load AWS config: %v", err)
		}
	}

	var client *s3.Client
	if endpoint != "" {
		client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		})
	} else {
		client = s3.NewFromConfig(cfg)
	}

	_, err = client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			log.Printf("Warning: failed to create bucket %s: %v", bucket, err)
		}
	}

	return &Storage{
		client: client,
		bucket: bucket,
	}
}

func (s *Storage) uploadKey(id string) string {
	return fmt.Sprintf("uploads/%s", id)
}

func (s *Storage) blobKey(d digest.Digest) string {
	return fmt.Sprintf("blobs/%s", d.String())
}

func (s *Storage) GetUploadBlobSize(id string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	key := s.uploadKey(id)
	output, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var notFound *types.NotFound
		if ok := isNotFoundError(err, notFound); ok {
			return 0, storage.ErrNotFound
		}
		log.Printf("failed to head object %s: %v", key, err)
		return 0, storage.ErrStorageFail
	}

	if output.ContentLength == nil {
		return 0, nil
	}
	return *output.ContentLength, nil
}

func (s *Storage) UploadBlob(id string, r io.Reader) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	key := s.uploadKey(id)

	existingData := make([]byte, 0)
	existingSize, err := s.GetUploadBlobSize(id)
	if err == nil && existingSize > 0 {
		getOutput, err := s.client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
		})
		if err == nil {
			existingData, _ = io.ReadAll(getOutput.Body)
			getOutput.Body.Close()
		}
	}

	newData, err := io.ReadAll(r)
	if err != nil {
		log.Printf("failed to read data: %v", err)
		return 0, fmt.Errorf("failed to read data: %w", storage.ErrStorageFail)
	}

	combinedData := append(existingData, newData...)

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(combinedData),
	})
	if err != nil {
		log.Printf("failed to upload blob %s: %v", key, err)
		return 0, fmt.Errorf("failed to upload blob: %w", storage.ErrStorageFail)
	}

	return int64(len(combinedData)), nil
}

func (s *Storage) CommitBlob(id string, d digest.Digest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	uploadKey := s.uploadKey(id)
	blobKey := s.blobKey(d)

	// Check if blob already exists
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(blobKey),
	})
	if err == nil {
		// Blob already exists, just delete the upload and return
		_, _ = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(uploadKey),
		})
		return nil
	}

	getOutput, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(uploadKey),
	})
	if err != nil {
		log.Printf("failed to get uploaded file %s: %v", uploadKey, err)
		return fmt.Errorf("failed to get uploaded file: %w", storage.ErrStorageFail)
	}
	defer getOutput.Body.Close()

	verifier := d.Verifier()
	data, err := io.ReadAll(io.TeeReader(getOutput.Body, verifier))
	if err != nil {
		log.Printf("failed to read uploaded file: %v", err)
		return fmt.Errorf("failed to read uploaded file: %w", storage.ErrStorageFail)
	}

	if !verifier.Verified() {
		log.Printf("digest not verified: %s", d.String())
		return storage.ErrNotVerified
	}

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(blobKey),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		log.Printf("failed to store blob %s: %v", blobKey, err)
		return fmt.Errorf("failed to store blob: %w", storage.ErrStorageFail)
	}

	_, _ = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(uploadKey),
	})

	return nil
}

func (s *Storage) SaveBlob(d digest.Digest, data []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	key := s.blobKey(d)

	// Check if blob already exists
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err == nil {
		// Already exists
		return nil
	}

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		log.Printf("failed to save blob %s: %v", key, err)
		return fmt.Errorf("failed to save blob: %w", storage.ErrStorageFail)
	}

	return nil
}

func (s *Storage) ReadBlob(d digest.Digest) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	key := s.blobKey(d)
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var notFound *types.NotFound
		if ok := isNotFoundError(err, notFound); ok {
			return nil, storage.ErrNotFound
		}
		log.Printf("failed to read blob %s: %v", key, err)
		return nil, fmt.Errorf("failed to read blob: %w", storage.ErrStorageFail)
	}
	defer output.Body.Close()

	data, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read blob body: %w", storage.ErrStorageFail)
	}

	return data, nil
}

func (s *Storage) ReadBlobToWriter(d digest.Digest, w io.Writer) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	key := s.blobKey(d)
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var notFound *types.NotFound
		if ok := isNotFoundError(err, notFound); ok {
			return 0, storage.ErrNotFound
		}
		log.Printf("failed to read blob %s: %v", key, err)
		return 0, fmt.Errorf("failed to read blob: %w", storage.ErrStorageFail)
	}
	defer output.Body.Close()

	verifier := d.Verifier()
	writer := io.MultiWriter(w, verifier)

	size, err := io.Copy(writer, output.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to copy blob: %w", storage.ErrStorageFail)
	}

	if !verifier.Verified() {
		return 0, storage.ErrNotVerified
	}

	return size, nil
}

func (s *Storage) IsExistBlob(d digest.Digest) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	key := s.blobKey(d)
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var notFound *types.NotFound
		if ok := isNotFoundError(err, notFound); ok {
			return false, nil
		}
		return false, storage.ErrStorageFail
	}
	return true, nil
}

func (s *Storage) DeleteBlob(d digest.Digest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	key := s.blobKey(d)

	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var notFound *types.NotFound
		if ok := isNotFoundError(err, notFound); ok {
			return storage.ErrNotFound
		}
		return storage.ErrStorageFail
	}

	_, err = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete blob: %w", storage.ErrStorageFail)
	}

	return nil
}

func isNotFoundError(err error, _ *types.NotFound) bool {
	if err == nil {
		return false
	}
	var noSuchKey *types.NoSuchKey
	var notFound *types.NotFound
	if ok := isErrorType(err, &noSuchKey); ok {
		return true
	}
	if ok := isErrorType(err, &notFound); ok {
		return true
	}
	errMsg := err.Error()
	return contains(errMsg, "NotFound") || contains(errMsg, "NoSuchKey") || contains(errMsg, "404")
}

func isErrorType[T any](err error, target *T) bool {
	return err != nil && target != nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
