package domain

import (
	"errors"
	"time"
)

type Instance struct {
	InstanceID  string
	UserID      string
	ChallengeID string
	ImageTag    string
	Status      InstanceStatus
	Host        string
	Port        int32
	StartedAt   time.Time
	ExpiresAt   time.Time
}

var (
	ErrInstanceNotFound      = errors.New("instance not found")
	ErrInstanceAlreadyExists = errors.New("instance already exists")
)

type InstanceStatus string

const (
	InstanceStatusRunning   InstanceStatus = "running"
	InstanceStatusStopped   InstanceStatus = "stopped"
	InstanceStatusDestroyed InstanceStatus = "destroyed"
	InstanceStatusUnknown   InstanceStatus = "unknown"
)
