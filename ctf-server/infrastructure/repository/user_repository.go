package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/kavos113/quickctf/ctf-server/domain"
)

type MySQLUserRepository struct {
	db *sql.DB
}

func NewMySQLUserRepository(db *sql.DB) *MySQLUserRepository {
	return &MySQLUserRepository{
		db: db,
	}
}

func (r *MySQLUserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, is_admin, created_at, updated_at)
		VALUES (?, ?, '', ?, FALSE, ?, ?)
	`
	
	_, err := r.db.ExecContext(ctx, query,
		user.UserID,
		user.Username,
		user.PasswordHash,
		user.CreatedAt,
		user.UpdatedAt,
	)
	
	if err != nil {
		return err
	}
	
	return nil
}

func (r *MySQLUserRepository) FindByID(ctx context.Context, userID string) (*domain.User, error) {
	query := `
		SELECT id, username, password_hash, created_at, updated_at
		FROM users
		WHERE id = ?
	`
	
	var user domain.User
	
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.UserID,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}

func (r *MySQLUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, username, password_hash, created_at, updated_at
		FROM users
		WHERE username = ?
	`
	
	var user domain.User
	
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.UserID,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}

func (r *MySQLUserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET username = ?, password_hash = ?, updated_at = ?
		WHERE id = ?
	`
	
	result, err := r.db.ExecContext(ctx, query,
		user.Username,
		user.PasswordHash,
		time.Now(),
		user.UserID,
	)
	
	if err != nil {
		return err
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rows == 0 {
		return domain.ErrUserNotFound
	}
	
	return nil
}

func (r *MySQLUserRepository) Delete(ctx context.Context, userID string) error {
	query := `DELETE FROM users WHERE id = ?`
	
	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rows == 0 {
		return domain.ErrUserNotFound
	}
	
	return nil
}
