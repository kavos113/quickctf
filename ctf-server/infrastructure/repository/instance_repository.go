package repository

import (
	"context"
	"database/sql"

	"github.com/kavos113/quickctf/ctf-server/domain"
)

type MySQLInstanceRepository struct {
	db *sql.DB
}

func NewMySQLInstanceRepository(db *sql.DB) *MySQLInstanceRepository {
	return &MySQLInstanceRepository{db: db}
}

func (r *MySQLInstanceRepository) Create(ctx context.Context, instance *domain.Instance) error {
	query := `
		INSERT INTO instances (id, user_id, challenge_id, image_tag, status, host, port, started_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		instance.InstanceID,
		instance.UserID,
		instance.ChallengeID,
		instance.ImageTag,
		instance.Status,
		instance.Host,
		instance.Port,
		instance.StartedAt,
		instance.ExpiresAt,
	)
	return err
}

func (r *MySQLInstanceRepository) FindByID(ctx context.Context, instanceID string) (*domain.Instance, error) {
	query := `
		SELECT id, user_id, challenge_id, image_tag, status, host, port, started_at, expires_at
		FROM instances
		WHERE id = ?
	`
	instance := &domain.Instance{}
	err := r.db.QueryRowContext(ctx, query, instanceID).Scan(
		&instance.InstanceID,
		&instance.UserID,
		&instance.ChallengeID,
		&instance.ImageTag,
		&instance.Status,
		&instance.Host,
		&instance.Port,
		&instance.StartedAt,
		&instance.ExpiresAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrInstanceNotFound
	}
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func (r *MySQLInstanceRepository) FindByUserAndChallenge(ctx context.Context, userID, challengeID string) (*domain.Instance, error) {
	query := `
		SELECT id, user_id, challenge_id, image_tag, status, host, port, started_at, expires_at
		FROM instances
		WHERE user_id = ? AND challenge_id = ? AND status != 'destroyed'
		ORDER BY started_at DESC
		LIMIT 1
	`
	instance := &domain.Instance{}
	err := r.db.QueryRowContext(ctx, query, userID, challengeID).Scan(
		&instance.InstanceID,
		&instance.UserID,
		&instance.ChallengeID,
		&instance.ImageTag,
		&instance.Status,
		&instance.Host,
		&instance.Port,
		&instance.StartedAt,
		&instance.ExpiresAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrInstanceNotFound
	}
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func (r *MySQLInstanceRepository) Update(ctx context.Context, instance *domain.Instance) error {
	query := `
		UPDATE instances
		SET status = ?, host = ?, port = ?
		WHERE id = ?
	`
	result, err := r.db.ExecContext(ctx, query,
		instance.Status,
		instance.Host,
		instance.Port,
		instance.InstanceID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrInstanceNotFound
	}

	return nil
}

func (r *MySQLInstanceRepository) Delete(ctx context.Context, instanceID string) error {
	query := `DELETE FROM instances WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, instanceID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrInstanceNotFound
	}

	return nil
}
