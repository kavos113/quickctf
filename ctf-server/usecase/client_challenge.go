package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kavos113/quickctf/ctf-server/domain"
)

type ClientChallengeUsecase struct {
	challengeRepo   domain.ChallengeRepository
	submissionRepo  domain.SubmissionRepository
}

func NewClientChallengeUsecase(
	challengeRepo domain.ChallengeRepository,
	submissionRepo domain.SubmissionRepository,
) *ClientChallengeUsecase {
	return &ClientChallengeUsecase{
		challengeRepo:   challengeRepo,
		submissionRepo:  submissionRepo,
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

func (u *ClientChallengeUsecase) StartInstance(ctx context.Context, userID, challengeID string) error {
	_, err := u.challengeRepo.FindByID(ctx, challengeID)
	if err != nil {
		return err
	}

	// TODO: ctf-managerを呼び出してインスタンスを起動
	return nil
}

func (u *ClientChallengeUsecase) StopInstance(ctx context.Context, userID, challengeID string) error {
	_, err := u.challengeRepo.FindByID(ctx, challengeID)
	if err != nil {
		return err
	}

	// TODO: ctf-managerを呼び出してインスタンスを停止
	return nil
}

func (u *ClientChallengeUsecase) GetInstanceStatus(ctx context.Context, userID, challengeID string) (string, error) {
	_, err := u.challengeRepo.FindByID(ctx, challengeID)
	if err != nil {
		return "", err
	}

	// TODO: ctf-managerを呼び出してインスタンスの状態を取得
	return "not_implemented", nil
}
