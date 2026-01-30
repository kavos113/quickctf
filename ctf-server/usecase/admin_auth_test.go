package usecase

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/kavos113/quickctf/ctf-server/domain"
)

func TestAdminAuthUsecase_ActivateAdmin(t *testing.T) {
	ctx := context.Background()
	sessionRepo := NewMockSessionRepository()

	testCode := "test_activation_code"
	os.Setenv("ADMIN_ACTIVATION_CODE", testCode)
	defer os.Unsetenv("ADMIN_ACTIVATION_CODE")

	uc := NewAdminAuthUsecase(sessionRepo)

	session := &domain.Session{
		SessionID: "test-session",
		UserID:    "test-user",
		Token:     "test-token",
		IsAdmin:   false,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	sessionRepo.Create(ctx, session)

	tests := []struct {
		name           string
		token          string
		activationCode string
		wantErr        bool
		wantIsAdmin    bool
	}{
		{
			name:           "valid activation code",
			token:          "test-token",
			activationCode: testCode,
			wantErr:        false,
			wantIsAdmin:    true,
		},
		{
			name:           "invalid activation code",
			token:          "test-token",
			activationCode: "wrong-code",
			wantErr:        true,
			wantIsAdmin:    false,
		},
		{
			name:           "invalid token",
			token:          "invalid-token",
			activationCode: testCode,
			wantErr:        true,
			wantIsAdmin:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session.IsAdmin = false
			sessionRepo.Update(ctx, session)

			err := uc.ActivateAdmin(ctx, tt.token, tt.activationCode)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ActivateAdmin() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ActivateAdmin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			updatedSession, _ := sessionRepo.FindByToken(ctx, tt.token)
			if updatedSession.IsAdmin != tt.wantIsAdmin {
				t.Errorf("ActivateAdmin() IsAdmin = %v, want %v", updatedSession.IsAdmin, tt.wantIsAdmin)
			}
		})
	}
}

func TestAdminAuthUsecase_ValidateAdminToken(t *testing.T) {
	ctx := context.Background()
	sessionRepo := NewMockSessionRepository()
	uc := NewAdminAuthUsecase(sessionRepo)

	adminSession := &domain.Session{
		SessionID: "admin-session",
		UserID:    "admin-user",
		Token:     "admin-token",
		IsAdmin:   true,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	regularSession := &domain.Session{
		SessionID: "regular-session",
		UserID:    "regular-user",
		Token:     "regular-token",
		IsAdmin:   false,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	sessionRepo.Create(ctx, adminSession)
	sessionRepo.Create(ctx, regularSession)

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid admin token",
			token:   "admin-token",
			wantErr: false,
		},
		{
			name:    "regular user token",
			token:   "regular-token",
			wantErr: true,
		},
		{
			name:    "invalid token",
			token:   "invalid-token",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := uc.ValidateAdminToken(ctx, tt.token)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateAdminToken() error = nil, wantErr %v", tt.wantErr)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateAdminToken() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestAdminAuthUsecase_DeactivateAdmin(t *testing.T) {
	ctx := context.Background()
	sessionRepo := NewMockSessionRepository()
	uc := NewAdminAuthUsecase(sessionRepo)

	adminSession := &domain.Session{
		SessionID: "admin-session",
		UserID:    "admin-user",
		Token:     "admin-token",
		IsAdmin:   true,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	sessionRepo.Create(ctx, adminSession)

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid deactivation",
			token:   "admin-token",
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
			adminSession.IsAdmin = true
			sessionRepo.Update(ctx, adminSession)

			err := uc.DeactivateAdmin(ctx, tt.token)

			if tt.wantErr {
				if err == nil {
					t.Errorf("DeactivateAdmin() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("DeactivateAdmin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			updatedSession, _ := sessionRepo.FindByToken(ctx, tt.token)
			if updatedSession.IsAdmin {
				t.Errorf("DeactivateAdmin() IsAdmin = true, want false")
			}
		})
	}
}
