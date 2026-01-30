package repository

import (
	"context"
	"database/sql"

	"github.com/kavos113/quickctf/ctf-server/domain"
)

type MySQLSubmissionRepository struct {
	db *sql.DB
}

func NewMySQLSubmissionRepository(db *sql.DB) *MySQLSubmissionRepository {
	return &MySQLSubmissionRepository{db: db}
}

func (r *MySQLSubmissionRepository) Create(ctx context.Context, submission *domain.Submission) error {
	query := `
		INSERT INTO submissions (id, user_id, challenge_id, submitted_flag, is_correct, submitted_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		submission.SubmissionID,
		submission.UserID,
		submission.ChallengeID,
		submission.SubmittedFlag,
		submission.IsCorrect,
		submission.SubmittedAt,
	)
	return err
}

func (r *MySQLSubmissionRepository) FindByID(ctx context.Context, submissionID string) (*domain.Submission, error) {
	query := `
		SELECT id, user_id, challenge_id, submitted_flag, is_correct, submitted_at
		FROM submissions
		WHERE id = ?
	`
	submission := &domain.Submission{}
	err := r.db.QueryRowContext(ctx, query, submissionID).Scan(
		&submission.SubmissionID,
		&submission.UserID,
		&submission.ChallengeID,
		&submission.SubmittedFlag,
		&submission.IsCorrect,
		&submission.SubmittedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrSubmissionNotFound
	}
	if err != nil {
		return nil, err
	}
	return submission, nil
}

func (r *MySQLSubmissionRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.Submission, error) {
	query := `
		SELECT id, user_id, challenge_id, submitted_flag, is_correct, submitted_at
		FROM submissions
		WHERE user_id = ?
		ORDER BY submitted_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var submissions []*domain.Submission
	for rows.Next() {
		submission := &domain.Submission{}
		if err := rows.Scan(
			&submission.SubmissionID,
			&submission.UserID,
			&submission.ChallengeID,
			&submission.SubmittedFlag,
			&submission.IsCorrect,
			&submission.SubmittedAt,
		); err != nil {
			return nil, err
		}
		submissions = append(submissions, submission)
	}

	return submissions, rows.Err()
}

func (r *MySQLSubmissionRepository) FindByChallengeID(ctx context.Context, challengeID string) ([]*domain.Submission, error) {
	query := `
		SELECT id, user_id, challenge_id, submitted_flag, is_correct, submitted_at
		FROM submissions
		WHERE challenge_id = ?
		ORDER BY submitted_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, challengeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var submissions []*domain.Submission
	for rows.Next() {
		submission := &domain.Submission{}
		if err := rows.Scan(
			&submission.SubmissionID,
			&submission.UserID,
			&submission.ChallengeID,
			&submission.SubmittedFlag,
			&submission.IsCorrect,
			&submission.SubmittedAt,
		); err != nil {
			return nil, err
		}
		submissions = append(submissions, submission)
	}

	return submissions, rows.Err()
}

func (r *MySQLSubmissionRepository) FindByUserAndChallenge(ctx context.Context, userID, challengeID string) ([]*domain.Submission, error) {
	query := `
		SELECT id, user_id, challenge_id, submitted_flag, is_correct, submitted_at
		FROM submissions
		WHERE user_id = ? AND challenge_id = ?
		ORDER BY submitted_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID, challengeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var submissions []*domain.Submission
	for rows.Next() {
		submission := &domain.Submission{}
		if err := rows.Scan(
			&submission.SubmissionID,
			&submission.UserID,
			&submission.ChallengeID,
			&submission.SubmittedFlag,
			&submission.IsCorrect,
			&submission.SubmittedAt,
		); err != nil {
			return nil, err
		}
		submissions = append(submissions, submission)
	}

	return submissions, rows.Err()
}
