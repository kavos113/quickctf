package queue

import (
	"encoding/json"
	"time"
)

type BuildJob struct {
	JobID      string    `json:"job_id"`
	ImageTag   string    `json:"image_tag"`
	SourceTar  []byte    `json:"source_tar"`
	CreatedAt  time.Time `json:"created_at"`
	ChallengeID string   `json:"challenge_id"`
}

type BuildResult struct {
	JobID        string    `json:"job_id"`
	ImageID      string    `json:"image_id"`
	Status       string    `json:"status"` // "success", "failed", "building"
	ErrorMessage string    `json:"error_message,omitempty"`
	CompletedAt  time.Time `json:"completed_at,omitempty"`
}

const (
	BuildStatusPending  = "pending"
	BuildStatusBuilding = "building"
	BuildStatusSuccess  = "success"
	BuildStatusFailed   = "failed"
)

func (j *BuildJob) ToJSON() ([]byte, error) {
	return json.Marshal(j)
}

func ParseBuildJob(data []byte) (*BuildJob, error) {
	var job BuildJob
	if err := json.Unmarshal(data, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *BuildResult) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

func ParseBuildResult(data []byte) (*BuildResult, error) {
	var result BuildResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
