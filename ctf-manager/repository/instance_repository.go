package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/kavos113/quickctf/ctf-manager/domain"
)

type MySQLInstanceRepository struct {
	db *sql.DB
}

func NewMySQLInstanceRepository(db *sql.DB) *MySQLInstanceRepository {
	return &MySQLInstanceRepository{
		db: db,
	}
}

func (r *MySQLInstanceRepository) Create(ctx context.Context, instance *domain.Instance) error {
	query := `
		INSERT INTO instances (instance_id, image_tag, runner_url, container_id, host, port, state, ttl_seconds, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		instance.InstanceID,
		instance.ImageTag,
		instance.RunnerURL,
		instance.ContainerID,
		instance.Host,
		instance.Port,
		string(instance.State),
		int64(instance.TTL.Seconds()),
		instance.CreatedAt,
		instance.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *MySQLInstanceRepository) FindByID(ctx context.Context, instanceID string) (*domain.Instance, error) {
	query := `
		SELECT instance_id, image_tag, runner_url, container_id, host, port, state, ttl_seconds, created_at, updated_at
		FROM instances
		WHERE instance_id = ?
	`

	var instance domain.Instance
	var ttlSeconds int64

	err := r.db.QueryRowContext(ctx, query, instanceID).Scan(
		&instance.InstanceID,
		&instance.ImageTag,
		&instance.RunnerURL,
		&instance.ContainerID,
		&instance.Host,
		&instance.Port,
		&instance.State,
		&ttlSeconds,
		&instance.CreatedAt,
		&instance.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrInstanceNotFound
	}
	if err != nil {
		return nil, err
	}

	instance.TTL = time.Duration(ttlSeconds) * time.Second

	return &instance, nil
}

func (r *MySQLInstanceRepository) Update(ctx context.Context, instance *domain.Instance) error {
	query := `
		UPDATE instances
		SET image_tag = ?, runner_url = ?, container_id = ?, host = ?, port = ?, state = ?, ttl_seconds = ?, updated_at = ?
		WHERE instance_id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		instance.ImageTag,
		instance.RunnerURL,
		instance.ContainerID,
		instance.Host,
		instance.Port,
		string(instance.State),
		int64(instance.TTL.Seconds()),
		instance.UpdatedAt,
		instance.InstanceID,
	)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrInstanceNotFound
	}

	return nil
}

func (r *MySQLInstanceRepository) Delete(ctx context.Context, instanceID string) error {
	query := `DELETE FROM instances WHERE instance_id = ?`

	result, err := r.db.ExecContext(ctx, query, instanceID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrInstanceNotFound
	}

	return nil
}

func (r *MySQLInstanceRepository) FindAll(ctx context.Context) ([]*domain.Instance, error) {
	query := `
		SELECT instance_id, image_tag, runner_url, container_id, host, port, state, ttl_seconds, created_at, updated_at
		FROM instances
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []*domain.Instance
	for rows.Next() {
		var instance domain.Instance
		var ttlSeconds int64

		err := rows.Scan(
			&instance.InstanceID,
			&instance.ImageTag,
			&instance.RunnerURL,
			&instance.ContainerID,
			&instance.Host,
			&instance.Port,
			&instance.State,
			&ttlSeconds,
			&instance.CreatedAt,
			&instance.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		instance.TTL = time.Duration(ttlSeconds) * time.Second
		instances = append(instances, &instance)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return instances, nil
}

func (r *MySQLInstanceRepository) FindByRunnerURL(ctx context.Context, runnerURL string) ([]*domain.Instance, error) {
	query := `
		SELECT instance_id, image_tag, runner_url, container_id, host, port, state, ttl_seconds, created_at, updated_at
		FROM instances
		WHERE runner_url = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, runnerURL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []*domain.Instance
	for rows.Next() {
		var instance domain.Instance
		var ttlSeconds int64

		err := rows.Scan(
			&instance.InstanceID,
			&instance.ImageTag,
			&instance.RunnerURL,
			&instance.ContainerID,
			&instance.Host,
			&instance.Port,
			&instance.State,
			&ttlSeconds,
			&instance.CreatedAt,
			&instance.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		instance.TTL = time.Duration(ttlSeconds) * time.Second
		instances = append(instances, &instance)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return instances, nil
}

func (r *MySQLInstanceRepository) FindExpired(ctx context.Context) ([]*domain.Instance, error) {
	query := `
		SELECT instance_id, image_tag, runner_url, container_id, host, port, state, ttl_seconds, created_at, updated_at
		FROM instances
		WHERE ttl_seconds > 0 AND TIMESTAMPDIFF(SECOND, created_at, NOW()) > ttl_seconds
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []*domain.Instance
	for rows.Next() {
		var instance domain.Instance
		var ttlSeconds int64

		err := rows.Scan(
			&instance.InstanceID,
			&instance.ImageTag,
			&instance.RunnerURL,
			&instance.ContainerID,
			&instance.Host,
			&instance.Port,
			&instance.State,
			&ttlSeconds,
			&instance.CreatedAt,
			&instance.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		instance.TTL = time.Duration(ttlSeconds) * time.Second
		instances = append(instances, &instance)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return instances, nil
}
