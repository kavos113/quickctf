package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/kavos113/quickctf/ctf-server/domain"
)

type MySQLChallengeRepository struct {
	db             *sql.DB
	attachmentRepo *AttachmentRepository
}

func NewMySQLChallengeRepository(db *sql.DB) *MySQLChallengeRepository {
	return &MySQLChallengeRepository{
		db:             db,
		attachmentRepo: NewAttachmentRepository(db),
	}
}

func (r *MySQLChallengeRepository) Create(ctx context.Context, challenge *domain.Challenge) error {
	query := `
		INSERT INTO challenges (id, name, description, flag, points, genre, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		challenge.ChallengeID,
		challenge.Name,
		challenge.Description,
		challenge.Flag,
		challenge.Points,
		challenge.Genre,
		now,
		now,
	)
	if err != nil {
		return err
	}
	challenge.CreatedAt = now
	challenge.UpdatedAt = now
	return nil
}

func (r *MySQLChallengeRepository) FindByID(ctx context.Context, challengeID string) (*domain.Challenge, error) {
	query := `
		SELECT id, name, description, flag, points, genre, created_at, updated_at
		FROM challenges
		WHERE id = ?
	`
	challenge := &domain.Challenge{}
	err := r.db.QueryRowContext(ctx, query, challengeID).Scan(
		&challenge.ChallengeID,
		&challenge.Name,
		&challenge.Description,
		&challenge.Flag,
		&challenge.Points,
		&challenge.Genre,
		&challenge.CreatedAt,
		&challenge.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrChallengeNotFound
	}
	if err != nil {
		return nil, err
	}

	attachments, err := r.attachmentRepo.FindByChallengeID(ctx, challengeID)
	if err != nil {
		return nil, err
	}
	challenge.Attachments = attachments

	return challenge, nil
}

func (r *MySQLChallengeRepository) FindAll(ctx context.Context) ([]*domain.Challenge, error) {
	query := `
		SELECT id, name, description, flag, points, genre, created_at, updated_at
		FROM challenges
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var challenges []*domain.Challenge
	for rows.Next() {
		challenge := &domain.Challenge{}
		if err := rows.Scan(
			&challenge.ChallengeID,
			&challenge.Name,
			&challenge.Description,
			&challenge.Flag,
			&challenge.Points,
			&challenge.Genre,
			&challenge.CreatedAt,
			&challenge.UpdatedAt,
		); err != nil {
			return nil, err
		}
		challenges = append(challenges, challenge)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, challenge := range challenges {
		attachments, err := r.attachmentRepo.FindByChallengeID(ctx, challenge.ChallengeID)
		if err != nil {
			return nil, err
		}
		challenge.Attachments = attachments
	}

	return challenges, nil
}

func (r *MySQLChallengeRepository) Update(ctx context.Context, challenge *domain.Challenge) error {
	query := `
		UPDATE challenges
		SET name = ?, description = ?, flag = ?, points = ?, genre = ?, updated_at = ?
		WHERE id = ?
	`
	now := time.Now()
	result, err := r.db.ExecContext(ctx, query,
		challenge.Name,
		challenge.Description,
		challenge.Flag,
		challenge.Points,
		challenge.Genre,
		now,
		challenge.ChallengeID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrChallengeNotFound
	}

	challenge.UpdatedAt = now
	return nil
}

func (r *MySQLChallengeRepository) Delete(ctx context.Context, challengeID string) error {
	query := `DELETE FROM challenges WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, challengeID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrChallengeNotFound
	}

	return nil
}
