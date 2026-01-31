package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kavos113/quickctf/ctf-server/domain"
	"github.com/kavos113/quickctf/ctf-server/infrastructure/client"
)

type AdminServiceUsecase struct {
	challengeRepo domain.ChallengeRepository
	builderClient *client.BuilderClient
}

func NewAdminServiceUsecase(challengeRepo domain.ChallengeRepository, sessionRepo domain.SessionRepository, builderClient *client.BuilderClient) *AdminServiceUsecase {
	return &AdminServiceUsecase{
		challengeRepo: challengeRepo,
		builderClient: builderClient,
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

	imageTag := fmt.Sprintf("ctf-%s:%s", challenge.Name, challengeID[:8])

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

	logContent, err := u.builderClient.GetBuildLog(ctx, jobID)
	if err != nil {
		return "", result.Status, fmt.Errorf("failed to get build log: %w", err)
	}

	return logContent, result.Status, nil
}
