package domain

import (
	"context"
	"errors"
	"time"
)

type Submission struct {
	SubmissionID  string
	UserID        string
	ChallengeID   string
	SubmittedFlag string
	IsCorrect     bool
	SubmittedAt   time.Time
}

var (
	ErrSubmissionNotFound = errors.New("submission not found")
	ErrIncorrectFlag      = errors.New("incorrect flag")
)

type SubmissionRepository interface {
	Create(ctx context.Context, submission *Submission) error
	FindByID(ctx context.Context, submissionID string) (*Submission, error)
	FindByUserID(ctx context.Context, userID string) ([]*Submission, error)
	FindByChallengeID(ctx context.Context, challengeID string) ([]*Submission, error)
	FindByUserAndChallenge(ctx context.Context, userID, challengeID string) ([]*Submission, error)
}
