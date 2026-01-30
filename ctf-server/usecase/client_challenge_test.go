package usecase

import (
	"context"
	"testing"

	"github.com/kavos113/quickctf/ctf-server/domain"
)

type MockChallengeRepository struct {
	challenges map[string]*domain.Challenge
}

func NewMockChallengeRepository() *MockChallengeRepository {
	return &MockChallengeRepository{
		challenges: make(map[string]*domain.Challenge),
	}
}

func (m *MockChallengeRepository) Create(ctx context.Context, challenge *domain.Challenge) error {
	if _, exists := m.challenges[challenge.ChallengeID]; exists {
		return domain.ErrChallengeAlreadyExists
	}
	m.challenges[challenge.ChallengeID] = challenge
	return nil
}

func (m *MockChallengeRepository) FindByID(ctx context.Context, challengeID string) (*domain.Challenge, error) {
	challenge, exists := m.challenges[challengeID]
	if !exists {
		return nil, domain.ErrChallengeNotFound
	}
	return challenge, nil
}

func (m *MockChallengeRepository) FindAll(ctx context.Context) ([]*domain.Challenge, error) {
	result := make([]*domain.Challenge, 0, len(m.challenges))
	for _, c := range m.challenges {
		result = append(result, c)
	}
	return result, nil
}

func (m *MockChallengeRepository) Update(ctx context.Context, challenge *domain.Challenge) error {
	if _, exists := m.challenges[challenge.ChallengeID]; !exists {
		return domain.ErrChallengeNotFound
	}
	m.challenges[challenge.ChallengeID] = challenge
	return nil
}

func (m *MockChallengeRepository) Delete(ctx context.Context, challengeID string) error {
	if _, exists := m.challenges[challengeID]; !exists {
		return domain.ErrChallengeNotFound
	}
	delete(m.challenges, challengeID)
	return nil
}

// MockSubmissionRepository is a mock implementation of domain.SubmissionRepository
type MockSubmissionRepository struct {
	submissions map[string]*domain.Submission
}

func NewMockSubmissionRepository() *MockSubmissionRepository {
	return &MockSubmissionRepository{
		submissions: make(map[string]*domain.Submission),
	}
}

func (m *MockSubmissionRepository) Create(ctx context.Context, submission *domain.Submission) error {
	m.submissions[submission.SubmissionID] = submission
	return nil
}

func (m *MockSubmissionRepository) FindByID(ctx context.Context, submissionID string) (*domain.Submission, error) {
	submission, exists := m.submissions[submissionID]
	if !exists {
		return nil, domain.ErrSubmissionNotFound
	}
	return submission, nil
}

func (m *MockSubmissionRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.Submission, error) {
	result := make([]*domain.Submission, 0)
	for _, s := range m.submissions {
		if s.UserID == userID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *MockSubmissionRepository) FindByChallengeID(ctx context.Context, challengeID string) ([]*domain.Submission, error) {
	result := make([]*domain.Submission, 0)
	for _, s := range m.submissions {
		if s.ChallengeID == challengeID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *MockSubmissionRepository) FindByUserAndChallenge(ctx context.Context, userID, challengeID string) ([]*domain.Submission, error) {
	result := make([]*domain.Submission, 0)
	for _, s := range m.submissions {
		if s.UserID == userID && s.ChallengeID == challengeID {
			result = append(result, s)
		}
	}
	return result, nil
}

func TestClientChallengeUsecase_GetChallenges(t *testing.T) {
	ctx := context.Background()
	challengeRepo := NewMockChallengeRepository()
	submissionRepo := NewMockSubmissionRepository()

	challenge1 := &domain.Challenge{
		ChallengeID: "1",
		Name:        "Challenge 1",
		Description: "Description 1",
		Flag:        "flag{test1}",
		Points:      100,
		Genre:       "web",
	}
	challenge2 := &domain.Challenge{
		ChallengeID: "2",
		Name:        "Challenge 2",
		Description: "Description 2",
		Flag:        "flag{test2}",
		Points:      200,
		Genre:       "crypto",
	}
	challengeRepo.Create(ctx, challenge1)
	challengeRepo.Create(ctx, challenge2)

	uc := &ClientChallengeUsecase{
		challengeRepo:  challengeRepo,
		submissionRepo: submissionRepo,
	}

	challenges, err := uc.GetChallenges(ctx)
	if err != nil {
		t.Fatalf("GetChallenges() error = %v", err)
	}

	if len(challenges) != 2 {
		t.Errorf("GetChallenges() returned %d challenges, want 2", len(challenges))
	}

	for _, c := range challenges {
		if c.Flag != "" {
			t.Errorf("GetChallenges() returned challenge with flag, want empty flag")
		}
	}
}

func TestClientChallengeUsecase_SubmitFlag(t *testing.T) {
	ctx := context.Background()
	challengeRepo := NewMockChallengeRepository()
	submissionRepo := NewMockSubmissionRepository()

	challenge := &domain.Challenge{
		ChallengeID: "1",
		Name:        "Test Challenge",
		Description: "Test Description",
		Flag:        "flag{correct}",
		Points:      100,
		Genre:       "web",
	}
	challengeRepo.Create(ctx, challenge)

	uc := &ClientChallengeUsecase{
		challengeRepo:  challengeRepo,
		submissionRepo: submissionRepo,
	}

	tests := []struct {
		name          string
		userID        string
		challengeID   string
		submittedFlag string
		wantCorrect   bool
		wantPoints    int
		wantErr       bool
	}{
		{
			name:          "correct flag",
			userID:        "user1",
			challengeID:   "1",
			submittedFlag: "flag{correct}",
			wantCorrect:   true,
			wantPoints:    100,
			wantErr:       false,
		},
		{
			name:          "incorrect flag",
			userID:        "user1",
			challengeID:   "1",
			submittedFlag: "flag{wrong}",
			wantCorrect:   false,
			wantPoints:    0,
			wantErr:       false,
		},
		{
			name:          "non-existent challenge",
			userID:        "user1",
			challengeID:   "999",
			submittedFlag: "flag{test}",
			wantCorrect:   false,
			wantPoints:    0,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isCorrect, pointsAwarded, err := uc.SubmitFlag(ctx, tt.userID, tt.challengeID, tt.submittedFlag)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SubmitFlag() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("SubmitFlag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if isCorrect != tt.wantCorrect {
				t.Errorf("SubmitFlag() isCorrect = %v, want %v", isCorrect, tt.wantCorrect)
			}

			if pointsAwarded != tt.wantPoints {
				t.Errorf("SubmitFlag() pointsAwarded = %v, want %v", pointsAwarded, tt.wantPoints)
			}
		})
	}
}

func TestClientChallengeUsecase_SubmitFlag_AlreadySolved(t *testing.T) {
	ctx := context.Background()
	challengeRepo := NewMockChallengeRepository()
	submissionRepo := NewMockSubmissionRepository()

	challenge := &domain.Challenge{
		ChallengeID: "1",
		Name:        "Test Challenge",
		Description: "Test Description",
		Flag:        "flag{correct}",
		Points:      100,
		Genre:       "web",
	}
	challengeRepo.Create(ctx, challenge)

	uc := &ClientChallengeUsecase{
		challengeRepo:  challengeRepo,
		submissionRepo: submissionRepo,
	}

	isCorrect, pointsAwarded, err := uc.SubmitFlag(ctx, "user1", "1", "flag{correct}")
	if err != nil {
		t.Fatalf("First submission failed: %v", err)
	}
	if !isCorrect || pointsAwarded != 100 {
		t.Fatalf("First submission should be correct with 100 points")
	}

	// Second submission (already solved)
	isCorrect, pointsAwarded, err = uc.SubmitFlag(ctx, "user1", "1", "flag{correct}")
	if err != nil {
		t.Fatalf("Second submission failed: %v", err)
	}
	if isCorrect || pointsAwarded != 0 {
		t.Errorf("Second submission should return false and 0 points (already solved)")
	}
}
