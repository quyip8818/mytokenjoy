package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"sms/backend/internal/domain/auth"
	"sms/backend/internal/domain/types"
)

// --- mock store ---

type mockStore struct {
	users    map[string]*types.User
	sessions map[string]*types.Session
}

var testUserID = uuid.Must(uuid.NewV7())

func newMockStore() *mockStore {
	hash, _ := bcrypt.GenerateFromPassword([]byte("pass123"), bcrypt.MinCost)
	return &mockStore{
		users: map[string]*types.User{
			"admin": {ID: testUserID, Username: "admin", PasswordHash: string(hash), RealName: "管理员", Role: "admin", Status: 1},
		},
		sessions: map[string]*types.Session{},
	}
}

func (m *mockStore) GetUserByUsername(_ context.Context, username string) (*types.User, error) {
	u, ok := m.users[username]
	if !ok {
		return nil, types.ErrNotFound
	}
	return u, nil
}

func (m *mockStore) GetUserByID(_ context.Context, id uuid.UUID) (*types.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, types.ErrNotFound
}

func (m *mockStore) CreateSession(_ context.Context, session *types.Session) error {
	m.sessions[session.Token] = session
	return nil
}

func (m *mockStore) GetSession(_ context.Context, token string) (*types.Session, error) {
	s, ok := m.sessions[token]
	if !ok {
		return nil, types.ErrNotFound
	}
	return s, nil
}

func (m *mockStore) DeleteSession(_ context.Context, token string) error {
	delete(m.sessions, token)
	return nil
}

func (m *mockStore) RotateSession(_ context.Context, oldToken, newToken string, expiresAt time.Time) error {
	s, ok := m.sessions[oldToken]
	if !ok {
		return types.ErrNotFound
	}
	delete(m.sessions, oldToken)
	s.Token = newToken
	s.ExpiresAt = expiresAt
	m.sessions[newToken] = s
	return nil
}

// --- tests ---

func newService() *auth.Service {
	return auth.NewService(newMockStore(), "test-secret", 15*time.Minute, 7*24*time.Hour)
}

func TestLogin_Success(t *testing.T) {
	t.Parallel()
	svc := newService()
	result, err := svc.Login(context.Background(), "admin", "pass123")
	if err != nil {
		t.Fatal(err)
	}
	if result.AccessToken == "" {
		t.Fatal("expected access token")
	}
	if result.RefreshToken == "" {
		t.Fatal("expected refresh token")
	}
	if result.User.Username != "admin" {
		t.Fatalf("expected admin, got %s", result.User.Username)
	}
	// password hash should be cleared
	if result.User.PasswordHash != "" {
		t.Fatal("password hash should be empty in result")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	t.Parallel()
	svc := newService()
	_, err := svc.Login(context.Background(), "admin", "wrong")
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}

func TestLogin_UnknownUser(t *testing.T) {
	t.Parallel()
	svc := newService()
	_, err := svc.Login(context.Background(), "nobody", "pass123")
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}

func TestRefresh_Success(t *testing.T) {
	t.Parallel()
	svc := newService()
	login, _ := svc.Login(context.Background(), "admin", "pass123")
	result, err := svc.Refresh(context.Background(), login.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if result.AccessToken == "" {
		t.Fatal("expected new access token")
	}
	if result.RefreshToken == login.RefreshToken {
		t.Fatal("expected rotated refresh token")
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	t.Parallel()
	svc := newService()
	_, err := svc.Refresh(context.Background(), "invalid-token")
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}

func TestLogout(t *testing.T) {
	t.Parallel()
	svc := newService()
	login, _ := svc.Login(context.Background(), "admin", "pass123")
	if err := svc.Logout(context.Background(), login.RefreshToken); err != nil {
		t.Fatal(err)
	}
	// refresh should fail after logout
	_, err := svc.Refresh(context.Background(), login.RefreshToken)
	if err == nil {
		t.Fatal("expected error after logout")
	}
}

func TestProfile(t *testing.T) {
	t.Parallel()
	svc := newService()
	user, err := svc.Profile(context.Background(), testUserID)
	if err != nil {
		t.Fatal(err)
	}
	if user.PasswordHash != "" {
		t.Fatal("password hash should be cleared")
	}
	if user.RealName != "管理员" {
		t.Fatalf("expected 管理员, got %s", user.RealName)
	}
}
