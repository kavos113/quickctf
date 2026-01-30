package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/kavos113/quickctf/ctf-server/domain"
	"golang.org/x/crypto/bcrypt"
)

type UserAuthUsecase struct {
	userRepo    domain.UserRepository
	sessionRepo domain.SessionRepository
}

func NewUserAuthUsecase(userRepo domain.UserRepository, sessionRepo domain.SessionRepository) *UserAuthUsecase {
	return &UserAuthUsecase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

func (u *UserAuthUsecase) Register(ctx context.Context, username, password string) (string, error) {
	_, err := u.userRepo.FindByUsername(ctx, username)
	if err == nil {
		return "", domain.ErrUserAlreadyExists
	}
	if err != domain.ErrUserNotFound {
		return "", err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	now := time.Now()
	user := &domain.User{
		UserID:       uuid.New().String(),
		Username:     username,
		PasswordHash: string(hashedPassword),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return "", err
	}

	return user.UserID, nil
}

func (u *UserAuthUsecase) Login(ctx context.Context, username, password string) (string, error) {
	user, err := u.userRepo.FindByUsername(ctx, username)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return "", domain.ErrInvalidPassword
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", domain.ErrInvalidPassword
	}

	token, err := generateToken()
	if err != nil {
		return "", err
	}

	session := &domain.Session{
		SessionID: uuid.New().String(),
		UserID:    user.UserID,
		Token:     token,
		IsAdmin:   false,
		ExpiresAt: time.Now().Add(24 * time.Hour), 
		CreatedAt: time.Now(),
	}

	if err := u.sessionRepo.Create(ctx, session); err != nil {
		return "", err
	}

	return token, nil
}

func (u *UserAuthUsecase) Logout(ctx context.Context, token string) error {
	return u.sessionRepo.Delete(ctx, token)
}

func (u *UserAuthUsecase) ValidateToken(ctx context.Context, token string) (string, error) {
	session, err := u.sessionRepo.FindByToken(ctx, token)
	if err != nil {
		return "", err
	}

	if session.IsExpired() {
		u.sessionRepo.Delete(ctx, token)
		return "", domain.ErrSessionExpired
	}

	return session.UserID, nil
}

func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
