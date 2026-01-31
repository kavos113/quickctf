package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/kavos113/quickctf/ctf-server/domain"
)

type AttachmentRepository struct {
	db *sql.DB
}

func NewAttachmentRepository(db *sql.DB) *AttachmentRepository {
	return &AttachmentRepository{db: db}
}

func (r *AttachmentRepository) Create(ctx context.Context, attachment *domain.Attachment) error {
	query := `
		INSERT INTO attachments (id, challenge_id, filename, s3_key, size, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		attachment.AttachmentID,
		attachment.ChallengeID,
		attachment.Filename,
		attachment.S3Key,
		attachment.Size,
		attachment.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *AttachmentRepository) FindByID(ctx context.Context, attachmentID string) (*domain.Attachment, error) {
	query := `
		SELECT id, challenge_id, filename, s3_key, size, created_at
		FROM attachments
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, attachmentID)

	var attachment domain.Attachment
	var createdAt []byte

	err := row.Scan(
		&attachment.AttachmentID,
		&attachment.ChallengeID,
		&attachment.Filename,
		&attachment.S3Key,
		&attachment.Size,
		&createdAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrAttachmentNotFound
	}
	if err != nil {
		return nil, err
	}

	attachment.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", string(createdAt))

	return &attachment, nil
}

func (r *AttachmentRepository) FindByChallengeID(ctx context.Context, challengeID string) ([]*domain.Attachment, error) {
	query := `
		SELECT id, challenge_id, filename, s3_key, size, created_at
		FROM attachments
		WHERE challenge_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, challengeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []*domain.Attachment
	for rows.Next() {
		var attachment domain.Attachment
		var createdAt []byte

		err := rows.Scan(
			&attachment.AttachmentID,
			&attachment.ChallengeID,
			&attachment.Filename,
			&attachment.S3Key,
			&attachment.Size,
			&createdAt,
		)
		if err != nil {
			return nil, err
		}

		attachment.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", string(createdAt))
		attachments = append(attachments, &attachment)
	}

	return attachments, nil
}

func (r *AttachmentRepository) Delete(ctx context.Context, attachmentID string) error {
	query := `DELETE FROM attachments WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, attachmentID)
	return err
}
