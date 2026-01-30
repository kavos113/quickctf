package domain

import (
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
	ErrChallengeNotFound     = errors.New("challenge not found")
	ErrChallengeAlreadyExists = errors.New("challenge already exists")
	ErrInvalidChallengeData  = errors.New("invalid challenge data")
)
