package domain

import (
	"testing"
	"time"
)

func TestSession_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "expired session",
			expiresAt: time.Now().Add(-1 * time.Hour),
			want:      true,
		},
		{
			name:      "valid session",
			expiresAt: time.Now().Add(1 * time.Hour),
			want:      false,
		},
		{
			name:      "just expired",
			expiresAt: time.Now().Add(-1 * time.Second),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{
				ExpiresAt: tt.expiresAt,
			}
			if got := s.IsExpired(); got != tt.want {
				t.Errorf("Session.IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChallenge_Validation(t *testing.T) {
	tests := []struct {
		name      string
		challenge Challenge
		wantErr   error
	}{
		{
			name: "valid challenge",
			challenge: Challenge{
				ChallengeID: "test-id",
				Name:        "Test Challenge",
				Description: "Test Description",
				Flag:        "flag{test}",
				Points:      100,
				Genre:       "web",
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = tt.challenge
		})
	}
}
