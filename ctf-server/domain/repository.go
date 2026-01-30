package domain

import "context"

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, userID string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, userID string) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	FindByToken(ctx context.Context, token string) (*Session, error)
	Update(ctx context.Context, session *Session) error
	Delete(ctx context.Context, token string) error
	DeleteByUserID(ctx context.Context, userID string) error
}

type ChallengeRepository interface {
	Create(ctx context.Context, challenge *Challenge) error
	FindByID(ctx context.Context, challengeID string) (*Challenge, error)
	FindAll(ctx context.Context) ([]*Challenge, error)
	Update(ctx context.Context, challenge *Challenge) error
	Delete(ctx context.Context, challengeID string) error
}
