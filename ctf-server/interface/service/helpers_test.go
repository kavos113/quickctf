package service

import (
	"context"
	"testing"

	"github.com/kavos113/quickctf/ctf-server/domain"
)

func TestGetUserIDFromContext(t *testing.T) {
	userID := "test-user-123"
	ctx := context.WithValue(context.Background(), userIDContextKey, userID)

	retrievedUserID, err := getUserIDFromContext(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if retrievedUserID != userID {
		t.Errorf("Expected user ID '%s', got '%s'", userID, retrievedUserID)
	}
}

func TestGetUserIDFromContext_NoUserID(t *testing.T) {
	ctx := context.Background()

	_, err := getUserIDFromContext(ctx)
	if err == nil {
		t.Error("Expected error for context without user ID")
	}
}

func TestGetSessionFromContext(t *testing.T) {
	session := &domain.Session{
		SessionID: "test-session",
		UserID:    "test-user",
		Token:     "test-token",
		IsAdmin:   false,
	}

	ctx := context.WithValue(context.Background(), sessionContextKey, session)

	retrievedSession, err := getSessionFromContext(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if retrievedSession == nil {
		t.Error("Expected session from context, got nil")
	}

	if retrievedSession.UserID != session.UserID {
		t.Errorf("Expected user ID '%s', got '%s'", session.UserID, retrievedSession.UserID)
	}

	if retrievedSession.Token != session.Token {
		t.Errorf("Expected token '%s', got '%s'", session.Token, retrievedSession.Token)
	}
}

func TestGetSessionFromContext_NoSession(t *testing.T) {
	ctx := context.Background()

	_, err := getSessionFromContext(ctx)
	if err == nil {
		t.Error("Expected error for context without session")
	}
}

func TestRequireAdminSession(t *testing.T) {
	tests := []struct {
		name      string
		session   *domain.Session
		wantErr   bool
	}{
		{
			name: "admin session",
			session: &domain.Session{
				SessionID: "admin-session",
				UserID:    "admin-user",
				Token:     "admin-token",
				IsAdmin:   true,
			},
			wantErr: false,
		},
		{
			name: "regular user session",
			session: &domain.Session{
				SessionID: "user-session",
				UserID:    "regular-user",
				Token:     "user-token",
				IsAdmin:   false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), sessionContextKey, tt.session)

			session, err := requireAdminSession(ctx)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error for non-admin session")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if session == nil {
				t.Error("Expected session, got nil")
			}

			if !session.IsAdmin {
				t.Error("Expected admin session")
			}
		})
	}
}

func TestRequireAdminSession_NoSession(t *testing.T) {
	ctx := context.Background()

	_, err := requireAdminSession(ctx)
	if err == nil {
		t.Error("Expected error for context without session")
	}
}

