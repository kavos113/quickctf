package domain

import (
	"context"
	"errors"
	"time"

	managerPb "github.com/kavos113/quickctf/gen/go/api/manager/v1"
)

type Instance struct {
	InstanceID string
	ImageTag   string
	RunnerURL  string
	Host       string
	Port       int32
	State      State
	TTL        time.Duration
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type State string

const (
	StateUnspecified State = "unspecified"
	StateRunning     State = "running"
	StateStopped     State = "stopped"
	StateDestroyed   State = "destroyed"
)

func (s State) ToProtoState() managerPb.GetInstanceStatusResponse_State {
	switch s {
	case StateRunning:
		return managerPb.GetInstanceStatusResponse_STATE_RUNNING
	case StateStopped:
		return managerPb.GetInstanceStatusResponse_STATE_STOPPED
	case StateDestroyed:
		return managerPb.GetInstanceStatusResponse_STATE_DESTROYED
	default:
		return managerPb.GetInstanceStatusResponse_STATE_UNSPECIFIED
	}
}

func FromProtoState(state managerPb.GetInstanceStatusResponse_State) State {
	switch state {
	case managerPb.GetInstanceStatusResponse_STATE_RUNNING:
		return StateRunning
	case managerPb.GetInstanceStatusResponse_STATE_STOPPED:
		return StateStopped
	case managerPb.GetInstanceStatusResponse_STATE_DESTROYED:
		return StateDestroyed
	default:
		return StateUnspecified
	}
}

func (i *Instance) GetConnectionInfo() *managerPb.ConnectionInfo {
	return &managerPb.ConnectionInfo{
		Host: i.Host,
		Port: i.Port,
	}
}

func (i *Instance) IsExpired() bool {
	if i.TTL == 0 {
		return false
	}
	return time.Since(i.CreatedAt) > i.TTL
}

func (i *Instance) UpdateState(state State) {
	i.State = state
	i.UpdatedAt = time.Now()
}

var (
	ErrInstanceNotFound      = errors.New("instance not found")
	ErrInstanceAlreadyExists = errors.New("instance already exists")
)

type InstanceRepository interface {
	Create(ctx context.Context, instance *Instance) error
	FindByID(ctx context.Context, instanceID string) (*Instance, error)
	Update(ctx context.Context, instance *Instance) error
	Delete(ctx context.Context, instanceID string) error
	FindAll(ctx context.Context) ([]*Instance, error)
	FindByRunnerURL(ctx context.Context, runnerURL string) ([]*Instance, error)
	FindExpired(ctx context.Context) ([]*Instance, error)
}
