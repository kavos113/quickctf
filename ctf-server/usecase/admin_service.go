package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/kavos113/quickctf/ctf-server/domain"
)

type AdminServiceUsecase struct {
	challengeRepo domain.ChallengeRepository
}

func NewAdminServiceUsecase(challengeRepo domain.ChallengeRepository, sessionRepo domain.SessionRepository) *AdminServiceUsecase {
	return &AdminServiceUsecase{
		challengeRepo: challengeRepo,
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
