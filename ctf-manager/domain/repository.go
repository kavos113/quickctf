package domain

import (
	"context"
	"errors"
)

var ErrInstanceNotFound = errors.New("instance not found")
var ErrInstanceAlreadyExists = errors.New("instance already exists")

type InstanceRepository interface {
	Create(ctx context.Context, instance *Instance) error
	FindByID(ctx context.Context, instanceID string) (*Instance, error)
	Update(ctx context.Context, instance *Instance) error
	Delete(ctx context.Context, instanceID string) error
	FindAll(ctx context.Context) ([]*Instance, error)
	FindByRunnerURL(ctx context.Context, runnerURL string) ([]*Instance, error)
	FindExpired(ctx context.Context) ([]*Instance, error)
}
