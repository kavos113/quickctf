package domain

import (
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
