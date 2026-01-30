package repository

import (
	"context"
	"database/sql"

	"github.com/kavos113/quickctf/ctf-server/domain"
)

type MySQLSessionRepository struct {
	db *sql.DB
}

func NewMySQLSessionRepository(db *sql.DB) *MySQLSessionRepository {
	return &MySQLSessionRepository{
		db: db,
	}
}

func (r *MySQLSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	query := `
		INSERT INTO sessions (id, user_id, token, is_admin, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	
	_, err := r.db.ExecContext(ctx, query,
		session.SessionID,
		session.UserID,
		session.Token,
		session.IsAdmin,
		session.ExpiresAt,
		session.CreatedAt,
	)
	
	return err
}

func (r *MySQLSessionRepository) FindByToken(ctx context.Context, token string) (*domain.Session, error) {
	query := `
		SELECT id, user_id, token, is_admin, expires_at, created_at
		FROM sessions
		WHERE token = ?
	`
	
	var session domain.Session
	
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&session.SessionID,
		&session.UserID,
		&session.Token,
		&session.IsAdmin,
		&session.ExpiresAt,
		&session.CreatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, domain.ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	
	return &session, nil
}

func (r *MySQLSessionRepository) Delete(ctx context.Context, token string) error {
	query := `DELETE FROM sessions WHERE token = ?`
	
	result, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return err
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rows == 0 {
		return domain.ErrSessionNotFound
	}
	
	return nil
}

func (r *MySQLSessionRepository) Update(ctx context.Context, session *domain.Session) error {
	query := `
		UPDATE sessions
		SET is_admin = ?
		WHERE id = ?
	`
	
	result, err := r.db.ExecContext(ctx, query,
		session.IsAdmin,
		session.SessionID,
	)
	
	if err != nil {
		return err
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rows == 0 {
		return domain.ErrSessionNotFound
	}
	
	return nil
}

func (r *MySQLSessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	query := `DELETE FROM sessions WHERE user_id = ?`
	
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
