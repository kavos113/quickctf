package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kavos113/quickctf/ctf-server/domain"
)

type MockUserRepository struct {
	users map[string]*domain.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*domain.User),
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	if _, exists := m.users[user.UserID]; exists {
		return domain.ErrUserAlreadyExists
	}
	m.users[user.UserID] = user
	return nil
}

func (m *MockUserRepository) FindByID(ctx context.Context, userID string) (*domain.User, error) {
	user, exists := m.users[userID]
	if !exists {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (m *MockUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	if _, exists := m.users[user.UserID]; !exists {
		return domain.ErrUserNotFound
	}
	m.users[user.UserID] = user
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, userID string) error {
	if _, exists := m.users[userID]; !exists {
		return domain.ErrUserNotFound
	}
	delete(m.users, userID)
	return nil
}

type MockSessionRepository struct {
	sessions map[string]*domain.Session
}

func NewMockSessionRepository() *MockSessionRepository {
	return &MockSessionRepository{
		sessions: make(map[string]*domain.Session),
	}
}

func (m *MockSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	m.sessions[session.Token] = session
	return nil
}

func (m *MockSessionRepository) FindByToken(ctx context.Context, token string) (*domain.Session, error) {
	session, exists := m.sessions[token]
	if !exists {
		return nil, domain.ErrSessionNotFound
	}
	return session, nil
}

func (m *MockSessionRepository) Update(ctx context.Context, session *domain.Session) error {
	if _, exists := m.sessions[session.Token]; !exists {
		return domain.ErrSessionNotFound
	}
	m.sessions[session.Token] = session
	return nil
}

func (m *MockSessionRepository) Delete(ctx context.Context, token string) error {
	if _, exists := m.sessions[token]; !exists {
		return domain.ErrSessionNotFound
	}
	delete(m.sessions, token)
	return nil
}

func (m *MockSessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	for token, session := range m.sessions {
		if session.UserID == userID {
			delete(m.sessions, token)
		}
	}
	return nil
}

func TestUserAuthUsecase_Register(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		wantErr  bool
	}{
		{
			name:     "valid registration",
			username: "testuser",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "empty username",
			username: "",
			password: "password123",
			wantErr:  true,
		},
		{
			name:     "empty password",
			username: "testuser",
			password: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := NewMockUserRepository()
			sessionRepo := NewMockSessionRepository()
			uc := NewUserAuthUsecase(userRepo, sessionRepo)

			ctx := context.Background()
			userID, err := uc.Register(ctx, tt.username, tt.password)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Register() error = nil, wantErr %v", tt.wantErr)
				}
			} else {
				if err != nil {
					t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
				}
				if userID == "" {
					t.Errorf("Register() returned empty userID")
				}
			}
		})
	}
}

func TestUserAuthUsecase_Login(t *testing.T) {
	ctx := context.Background()
	userRepo := NewMockUserRepository()
	sessionRepo := NewMockSessionRepository()
	uc := NewUserAuthUsecase(userRepo, sessionRepo)

	username := "testuser"
	password := "password123"
	_, err := uc.Register(ctx, username, password)
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	tests := []struct {
		name     string
		username string
		password string
		wantErr  bool
	}{
		{
			name:     "valid login",
			username: username,
			password: password,
			wantErr:  false,
		},
		{
			name:     "invalid password",
			username: username,
			password: "wrongpassword",
			wantErr:  true,
		},
		{
			name:     "non-existent user",
			username: "nonexistent",
			password: password,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := uc.Login(ctx, tt.username, tt.password)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Login() error = nil, wantErr %v", tt.wantErr)
				}
			} else {
				if err != nil {
					t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
				}
				if token == "" {
					t.Errorf("Login() returned empty token")
				}
			}
		})
	}
}

func TestUserAuthUsecase_Logout(t *testing.T) {
	ctx := context.Background()
	userRepo := NewMockUserRepository()
	sessionRepo := NewMockSessionRepository()
	uc := NewUserAuthUsecase(userRepo, sessionRepo)

	username := "testuser"
	password := "password123"
	_, err := uc.Register(ctx, username, password)
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	token, err := uc.Login(ctx, username, password)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid logout",
			token:   token,
			wantErr: false,
		},
		{
			name:    "invalid token",
			token:   "invalid-token",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := uc.Logout(ctx, tt.token)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Logout() error = nil, wantErr %v", tt.wantErr)
				}
			} else {
				if err != nil {
					t.Errorf("Logout() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestUserAuthUsecase_ValidateToken(t *testing.T) {
	ctx := context.Background()
	userRepo := NewMockUserRepository()
	sessionRepo := NewMockSessionRepository()
	uc := NewUserAuthUsecase(userRepo, sessionRepo)

	username := "testuser"
	password := "password123"
	_, err := uc.Register(ctx, username, password)
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	token, err := uc.Login(ctx, username, password)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	expiredSession := &domain.Session{
		Token:     "expired-token",
		UserID:    "test-user",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	sessionRepo.Create(ctx, expiredSession)

	tests := []struct {
		name    string
		token   string
		wantErr error
	}{
		{
			name:    "valid token",
			token:   token,
			wantErr: nil,
		},
		{
			name:    "invalid token",
			token:   "invalid-token",
			wantErr: domain.ErrSessionNotFound,
		},
		{
			name:    "expired token",
			token:   "expired-token",
			wantErr: domain.ErrSessionExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.ValidateToken(ctx, tt.token)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("ValidateToken() error = nil, want %v", tt.wantErr)
				} else if !errors.Is(err, tt.wantErr) {
					t.Errorf("ValidateToken() error = %v, want %v", err, tt.wantErr)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateToken() error = %v, wantErr nil", err)
				}
			}
		})
	}
}
