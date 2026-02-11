package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kavos113/quickctf/ctf-server/domain"
	"github.com/kavos113/quickctf/ctf-server/infrastructure/client"
	"github.com/kavos113/quickctf/ctf-server/infrastructure/storage"
)

type AdminServiceUsecase struct {
	challengeRepo     domain.ChallengeRepository
	attachmentRepo    domain.AttachmentRepository
	builderClient     *client.BuilderClient
	attachmentStorage *storage.AttachmentStorage
	buildLogStorage   *storage.BuildLogStorage
}

func NewAdminServiceUsecase(
	challengeRepo domain.ChallengeRepository,
	attachmentRepo domain.AttachmentRepository,
	sessionRepo domain.SessionRepository,
	builderClient *client.BuilderClient,
	attachmentStorage *storage.AttachmentStorage,
	buildLogStorage *storage.BuildLogStorage,
) *AdminServiceUsecase {
	return &AdminServiceUsecase{
		challengeRepo:     challengeRepo,
		attachmentRepo:    attachmentRepo,
		builderClient:     builderClient,
		attachmentStorage: attachmentStorage,
		buildLogStorage:   buildLogStorage,
	}
}

func (u *AdminServiceUsecase) CreateChallenge(ctx context.Context, challenge *domain.Challenge) (string, error) {
	challenge.ChallengeID = uuid.New().String()
	if err := u.challengeRepo.Create(ctx, challenge); err != nil {
		return "", err
	}

	return challenge.ChallengeID, nil
}

func (u *AdminServiceUsecase) UpdateChallenge(ctx context.Context, challengeID string, challenge *domain.Challenge) error {
	challenge.ChallengeID = challengeID
	if err := u.challengeRepo.Update(ctx, challenge); err != nil {
		return err
	}

	return nil
}

func (u *AdminServiceUsecase) DeleteChallenge(ctx context.Context, challengeID string) error {
	if err := u.challengeRepo.Delete(ctx, challengeID); err != nil {
		return err
	}

	return nil
}

func (u *AdminServiceUsecase) ListChallenges(ctx context.Context) ([]*domain.Challenge, error) {
	challenges, err := u.challengeRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	return challenges, nil
}

func (u *AdminServiceUsecase) UploadChallengeImage(ctx context.Context, challengeID string, imageTar []byte) (string, error) {
	challenge, err := u.challengeRepo.FindByID(ctx, challengeID)
	if err != nil {
		return "", err
	}

	if !challenge.RequiresInstance {
		return "", fmt.Errorf("challenge does not require instance")
	}

	// TODO: challenge image version
	imageTag := fmt.Sprintf("ctf-%s:latest", challengeID)

	jobID, err := u.builderClient.BuildImage(ctx, imageTag, imageTar, challengeID)
	if err != nil {
		return "", fmt.Errorf("failed to build image: %w", err)
	}

	return jobID, nil
}

func (u *AdminServiceUsecase) GetChallenge(ctx context.Context, challengeID string) (*domain.Challenge, error) {
	return u.challengeRepo.FindByID(ctx, challengeID)
}

func (u *AdminServiceUsecase) ListBuildLogs(ctx context.Context, challengeID string) ([]client.BuildLogSummary, error) {
	return u.builderClient.ListBuildLogs(ctx, challengeID)
}

func (u *AdminServiceUsecase) GetBuildLog(ctx context.Context, jobID string) (string, string, error) {
	result, err := u.builderClient.GetBuildResult(ctx, jobID)
	if err != nil {
		return "", "", fmt.Errorf("failed to get build result: %w", err)
	}
	if result == nil {
		return "", "", fmt.Errorf("build job not found")
	}

	logContent, err := u.buildLogStorage.GetLog(ctx, jobID)
	if err != nil {
		return "", result.Status, fmt.Errorf("failed to get build log: %w", err)
	}

	return logContent, result.Status, nil
}

func (u *AdminServiceUsecase) GetBuildStatus(ctx context.Context, jobID string) (string, error) {
	result, err := u.builderClient.GetBuildResult(ctx, jobID)
	if err != nil {
		return "", fmt.Errorf("failed to get build result: %w", err)
	}
	if result == nil {
		return "", fmt.Errorf("build job not found")
	}
	return result.Status, nil
}

func (u *AdminServiceUsecase) SubscribeBuildLogs(ctx context.Context, jobID string, callback func(logLine string)) error {
	return u.builderClient.SubscribeBuildLogs(ctx, jobID, callback)
}

func (u *AdminServiceUsecase) UploadAttachment(ctx context.Context, challengeID string, filename string, data []byte) (*domain.Attachment, error) {
	_, err := u.challengeRepo.FindByID(ctx, challengeID)
	if err != nil {
		return nil, err
	}

	s3Key, err := u.attachmentStorage.Upload(ctx, challengeID, filename, data)
	if err != nil {
		return nil, fmt.Errorf("failed to upload attachment: %w", err)
	}

	attachment := &domain.Attachment{
		AttachmentID: uuid.New().String(),
		ChallengeID:  challengeID,
		Filename:     filename,
		S3Key:        s3Key,
		Size:         int64(len(data)),
		CreatedAt:    time.Now(),
	}

	if err := u.attachmentRepo.Create(ctx, attachment); err != nil {
		_ = u.attachmentStorage.Delete(ctx, s3Key)
		return nil, fmt.Errorf("failed to save attachment record: %w", err)
	}

	return attachment, nil
}

func (u *AdminServiceUsecase) DeleteAttachment(ctx context.Context, challengeID, attachmentID string) error {
	attachment, err := u.attachmentRepo.FindByID(ctx, attachmentID)
	if err != nil {
		return err
	}

	if attachment.ChallengeID != challengeID {
		return domain.ErrAttachmentNotFound
	}

	if err := u.attachmentStorage.Delete(ctx, attachment.S3Key); err != nil {
		return fmt.Errorf("failed to delete attachment from storage: %w", err)
	}

	if err := u.attachmentRepo.Delete(ctx, attachmentID); err != nil {
		return fmt.Errorf("failed to delete attachment record: %w", err)
	}

	return nil
}

func (u *AdminServiceUsecase) GetAttachmentURL(ctx context.Context, attachmentID string) (string, error) {
	attachment, err := u.attachmentRepo.FindByID(ctx, attachmentID)
	if err != nil {
		return "", err
	}

	url, err := u.attachmentStorage.GetPresignedURL(ctx, attachment.S3Key)
	if err != nil {
		return "", fmt.Errorf("failed to get presigned URL: %w", err)
	}

	return url, nil
}
