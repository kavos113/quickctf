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
	Attachments []*Attachment
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Attachment struct {
	AttachmentID string
	ChallengeID  string
	Filename     string
	S3Key        string
	Size         int64
	CreatedAt    time.Time
}

var (
	ErrChallengeNotFound      = errors.New("challenge not found")
	ErrChallengeAlreadyExists = errors.New("challenge already exists")
	ErrInvalidChallengeData   = errors.New("invalid challenge data")
	ErrAttachmentNotFound     = errors.New("attachment not found")
)

type ChallengeRepository interface {
	Create(ctx context.Context, challenge *Challenge) error
	FindByID(ctx context.Context, challengeID string) (*Challenge, error)
	FindAll(ctx context.Context) ([]*Challenge, error)
	Update(ctx context.Context, challenge *Challenge) error
	Delete(ctx context.Context, challengeID string) error
}

type AttachmentRepository interface {
	Create(ctx context.Context, attachment *Attachment) error
	FindByID(ctx context.Context, attachmentID string) (*Attachment, error)
	FindByChallengeID(ctx context.Context, challengeID string) ([]*Attachment, error)
	Delete(ctx context.Context, attachmentID string) error
}
