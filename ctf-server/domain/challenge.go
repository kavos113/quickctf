package domain

import (
	"context"
	"errors"
	"time"
)

type Challenge struct {
	ChallengeID string
	Name        string
	Description string
	Flag        string
	Points      int
	Genre       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

var (
	ErrChallengeNotFound      = errors.New("challenge not found")
	ErrChallengeAlreadyExists = errors.New("challenge already exists")
	ErrInvalidChallengeData   = errors.New("invalid challenge data")
)

type ChallengeRepository interface {
	Create(ctx context.Context, challenge *Challenge) error
	FindByID(ctx context.Context, challengeID string) (*Challenge, error)
	FindAll(ctx context.Context) ([]*Challenge, error)
	Update(ctx context.Context, challenge *Challenge) error
	Delete(ctx context.Context, challengeID string) error
}
