package domain

import (
	"errors"
	"time"
)

type User struct {
	UserID       string
	Username     string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Session struct {
	SessionID string
	UserID    string
	Token     string
	IsAdmin   bool
	ExpiresAt time.Time
	CreatedAt time.Time
}

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidPassword = errors.New("invalid password")
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired = errors.New("session expired")
	ErrInvalidActivationCode = errors.New("invalid activation code")
)

func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
