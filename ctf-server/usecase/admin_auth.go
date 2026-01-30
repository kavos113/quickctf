package usecase

import (
	"context"
	"os"

	"github.com/kavos113/quickctf/ctf-server/domain"
)

type AdminAuthUsecase struct {
	sessionRepo    domain.SessionRepository
	activationCode string
}

func NewAdminAuthUsecase(sessionRepo domain.SessionRepository) *AdminAuthUsecase {
	code := os.Getenv("ADMIN_ACTIVATION_CODE")
	if code == "" {
		code = "admin_secret"
	}

	return &AdminAuthUsecase{
		sessionRepo:    sessionRepo,
		activationCode: code,
	}
}

func (u *AdminAuthUsecase) ActivateAdmin(ctx context.Context, token, activationCode string) error {
	if activationCode != u.activationCode {
		return domain.ErrInvalidActivationCode
	}

	session, err := u.sessionRepo.FindByToken(ctx, token)
	if err != nil {
		return err
	}

	if session.IsExpired() {
		u.sessionRepo.Delete(ctx, token)
		return domain.ErrSessionExpired
	}

	if session.IsAdmin {
		return nil
	}

	session.IsAdmin = true
	if err := u.sessionRepo.Update(ctx, session); err != nil {
		return err
	}

	return nil
}

func (u *AdminAuthUsecase) ValidateAdminToken(ctx context.Context, token string) error {
	session, err := u.sessionRepo.FindByToken(ctx, token)
	if err != nil {
		return err
	}

	if session.IsExpired() {
		u.sessionRepo.Delete(ctx, token)
		return domain.ErrSessionExpired
	}

	if !session.IsAdmin {
		return domain.ErrInvalidActivationCode
	}

	return nil
}

func (u *AdminAuthUsecase) DeactivateAdmin(ctx context.Context, token string) error {
	session, err := u.sessionRepo.FindByToken(ctx, token)
	if err != nil {
		return err
	}

	if !session.IsAdmin {
		return nil
	}

	session.IsAdmin = false
	return u.sessionRepo.Update(ctx, session)
}
