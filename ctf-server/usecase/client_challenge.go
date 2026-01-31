package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kavos113/quickctf/ctf-server/domain"
	"github.com/kavos113/quickctf/ctf-server/infrastructure/client"
)

type ClientChallengeUsecase struct {
	challengeRepo  domain.ChallengeRepository
	submissionRepo domain.SubmissionRepository
	instanceRepo   domain.InstanceRepository
	managerClient  *client.ManagerClient
}

func NewClientChallengeUsecase(
	challengeRepo domain.ChallengeRepository,
	submissionRepo domain.SubmissionRepository,
	instanceRepo domain.InstanceRepository,
	managerClient *client.ManagerClient,
) *ClientChallengeUsecase {
	return &ClientChallengeUsecase{
		challengeRepo:  challengeRepo,
		submissionRepo: submissionRepo,
		instanceRepo:   instanceRepo,
		managerClient:  managerClient,
	}
}

func (u *ClientChallengeUsecase) GetChallenges(ctx context.Context) ([]*domain.Challenge, error) {
	challenges, err := u.challengeRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	for _, c := range challenges {
		c.Flag = ""
	}

	return challenges, nil
}

func (u *ClientChallengeUsecase) SubmitFlag(ctx context.Context, userID, challengeID, submittedFlag string) (bool, int, error) {
	challenge, err := u.challengeRepo.FindByID(ctx, challengeID)
	if err != nil {
		return false, 0, err
	}

	previousSubmissions, err := u.submissionRepo.FindByUserAndChallenge(ctx, userID, challengeID)
	if err != nil {
		return false, 0, err
	}

	for _, sub := range previousSubmissions {
		if sub.IsCorrect {
			return false, 0, nil
		}
	}

	isCorrect := challenge.Flag == submittedFlag

	submission := &domain.Submission{
		SubmissionID:  uuid.New().String(),
		UserID:        userID,
		ChallengeID:   challengeID,
		SubmittedFlag: submittedFlag,
		IsCorrect:     isCorrect,
		SubmittedAt:   time.Now(),
	}

	if err := u.submissionRepo.Create(ctx, submission); err != nil {
		return false, 0, err
	}

	pointsAwarded := 0
	if isCorrect {
		pointsAwarded = challenge.Points
	}

	return isCorrect, pointsAwarded, nil
}

func (u *ClientChallengeUsecase) StartInstance(ctx context.Context, userID, challengeID string) (string, int32, error) {
	challenge, err := u.challengeRepo.FindByID(ctx, challengeID)
	if err != nil {
		return "", 0, err
	}

	existingInstance, err := u.instanceRepo.FindByUserAndChallenge(ctx, userID, challengeID)
	if err == nil {
		if existingInstance.Status == domain.InstanceStatusRunning {
			return "", 0, fmt.Errorf("instance already running")
		}

		// 停止中のインスタンスがある場合は再起動
		if existingInstance.Status == domain.InstanceStatusStopped {
			ttlSeconds := int64(3600)
			connInfo, err := u.managerClient.StartInstance(ctx, existingInstance.ImageTag, existingInstance.InstanceID, ttlSeconds)
			if err != nil {
				return "", 0, fmt.Errorf("failed to restart instance: %w", err)
			}

			existingInstance.Status = domain.InstanceStatusRunning
			existingInstance.Host = connInfo.Host
			existingInstance.Port = connInfo.Port
			existingInstance.ExpiresAt = time.Now().Add(time.Duration(ttlSeconds) * time.Second)

			if err := u.instanceRepo.Update(ctx, existingInstance); err != nil {
				return "", 0, fmt.Errorf("failed to update instance: %w", err)
			}

			return connInfo.Host, connInfo.Port, nil
		}
	}

	instanceID := fmt.Sprintf("%s-%s", userID[:8], uuid.New().String()[:8])
	imageTag := fmt.Sprintf("ctf-%s:%s", challenge.Name, challenge.ChallengeID[:8])
	ttlSeconds := int64(3600)

	connInfo, err := u.managerClient.StartInstance(ctx, imageTag, instanceID, ttlSeconds)
	if err != nil {
		return "", 0, fmt.Errorf("failed to start instance: %w", err)
	}

	instance := &domain.Instance{
		InstanceID:  instanceID,
		UserID:      userID,
		ChallengeID: challengeID,
		ImageTag:    imageTag,
		Status:      domain.InstanceStatusRunning,
		Host:        connInfo.Host,
		Port:        connInfo.Port,
		StartedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(time.Duration(ttlSeconds) * time.Second),
	}

	if err := u.instanceRepo.Create(ctx, instance); err != nil {
		return "", 0, fmt.Errorf("failed to save instance: %w", err)
	}

	return connInfo.Host, connInfo.Port, nil
}

func (u *ClientChallengeUsecase) StopInstance(ctx context.Context, userID, challengeID string) error {
	_, err := u.challengeRepo.FindByID(ctx, challengeID)
	if err != nil {
		return err
	}

	instance, err := u.instanceRepo.FindByUserAndChallenge(ctx, userID, challengeID)
	if err != nil {
		return fmt.Errorf("instance not found: %w", err)
	}

	if instance.Status == domain.InstanceStatusStopped || instance.Status == domain.InstanceStatusDestroyed {
		return fmt.Errorf("instance already stopped")
	}

	if err := u.managerClient.StopInstance(ctx, instance.InstanceID); err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}

	instance.Status = domain.InstanceStatusStopped
	if err := u.instanceRepo.Update(ctx, instance); err != nil {
		return fmt.Errorf("failed to update instance: %w", err)
	}

	return nil
}

func (u *ClientChallengeUsecase) GetInstanceStatus(ctx context.Context, userID, challengeID string) (domain.InstanceStatus, string, int32, error) {
	_, err := u.challengeRepo.FindByID(ctx, challengeID)
	if err != nil {
		return domain.InstanceStatusUnknown, "", 0, err
	}

	instance, err := u.instanceRepo.FindByUserAndChallenge(ctx, userID, challengeID)
	if err != nil {
		if err == domain.ErrInstanceNotFound {
			return domain.InstanceStatusUnknown, "", 0, nil
		}
		return domain.InstanceStatusUnknown, "", 0, fmt.Errorf("failed to get instance: %w", err)
	}

	status, err := u.managerClient.GetInstanceStatus(ctx, instance.InstanceID)
	if err != nil {
		return instance.Status, instance.Host, instance.Port, nil
	}

	if status == domain.InstanceStatusDestroyed {
		if err := u.instanceRepo.Delete(ctx, instance.InstanceID); err != nil {
			return domain.InstanceStatusUnknown, "", 0, nil
		}
		return domain.InstanceStatusUnknown, "", 0, nil
	}

	if status != instance.Status {
		instance.Status = status
		u.instanceRepo.Update(ctx, instance)
	}

	return status, instance.Host, instance.Port, nil
}
