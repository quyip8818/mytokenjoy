package user_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"sms/backend/internal/domain/types"
	"sms/backend/internal/domain/user"
)

// --- mock store ---

type mockStore struct {
	users map[uuid.UUID]*types.User
}

func newMockStore() *mockStore {
	return &mockStore{
		users: map[uuid.UUID]*types.User{},
	}
}

func (m *mockStore) ListUsers(_ context.Context) ([]types.User, error) {
	out := make([]types.User, 0, len(m.users))
	for _, u := range m.users {
		out = append(out, *u)
	}
	return out, nil
}

func (m *mockStore) GetUserByID(_ context.Context, id uuid.UUID) (*types.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, types.ErrNotFound
	}
	return u, nil
}

func (m *mockStore) CreateUser(_ context.Context, u *types.User) error {
	if _, exists := m.users[u.ID]; exists {
		return types.ErrConflict
	}
	m.users[u.ID] = u
	return nil
}

func (m *mockStore) UpdateUser(_ context.Context, u *types.User) error {
	if _, ok := m.users[u.ID]; !ok {
		return types.ErrNotFound
	}
	m.users[u.ID] = u
	return nil
}

func (m *mockStore) DeleteUser(_ context.Context, id uuid.UUID) error {
	if _, ok := m.users[id]; !ok {
		return types.ErrNotFound
	}
	delete(m.users, id)
	return nil
}

// --- tests ---

func newService() *user.Service {
	return user.NewService(newMockStore())
}

func TestCreate_Success(t *testing.T) {
	t.Parallel()
	svc := newService()
	u, err := svc.Create(context.Background(), user.CreateInput{
		Username: "alice", Password: "secret123", RealName: "Alice",
	})
	if err != nil {
		t.Fatal(err)
	}
	if u.ID == uuid.Nil {
		t.Fatal("expected non-nil ID")
	}
	if u.Username != "alice" {
		t.Fatalf("expected username alice, got %s", u.Username)
	}
	if u.PasswordHash != "" {
		t.Fatal("password hash should be cleared in result")
	}
	if u.Role != "viewer" {
		t.Fatalf("expected default role viewer, got %s", u.Role)
	}
	if u.Status != 1 {
		t.Fatalf("expected default status 1, got %d", u.Status)
	}
}

func TestCreate_WithRole(t *testing.T) {
	t.Parallel()
	svc := newService()
	u, err := svc.Create(context.Background(), user.CreateInput{
		Username: "bob", Password: "pass", RealName: "Bob", Role: "admin",
	})
	if err != nil {
		t.Fatal(err)
	}
	if u.Role != "admin" {
		t.Fatalf("expected role admin, got %s", u.Role)
	}
}

func TestCreate_ValidationError(t *testing.T) {
	t.Parallel()
	svc := newService()
	cases := []struct {
		name  string
		input user.CreateInput
	}{
		{"empty username", user.CreateInput{Password: "x", RealName: "x"}},
		{"empty password", user.CreateInput{Username: "x", RealName: "x"}},
		{"empty realname", user.CreateInput{Username: "x", Password: "x"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), tc.input)
			if err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestUpdate_Success(t *testing.T) {
	t.Parallel()
	svc := newService()
	created, _ := svc.Create(context.Background(), user.CreateInput{
		Username: "carl", Password: "pass", RealName: "Carl",
	})
	updated, err := svc.Update(context.Background(), created.ID, user.UpdateInput{
		RealName: "Carl Updated", Role: "admin",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.RealName != "Carl Updated" {
		t.Fatalf("expected updated name, got %s", updated.RealName)
	}
	if updated.Role != "admin" {
		t.Fatalf("expected role admin, got %s", updated.Role)
	}
	if updated.PasswordHash != "" {
		t.Fatal("password hash should be cleared")
	}
}

func TestUpdate_ChangePassword(t *testing.T) {
	t.Parallel()
	svc := newService()
	created, _ := svc.Create(context.Background(), user.CreateInput{
		Username: "dan", Password: "old", RealName: "Dan",
	})
	_, err := svc.Update(context.Background(), created.ID, user.UpdateInput{
		RealName: "Dan", Password: "newpass",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	t.Parallel()
	svc := newService()
	_, err := svc.Update(context.Background(), uuid.Must(uuid.NewV7()), user.UpdateInput{RealName: "X"})
	if err == nil {
		t.Fatal("expected not found error")
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()
	svc := newService()
	created, _ := svc.Create(context.Background(), user.CreateInput{
		Username: "del", Password: "pass", RealName: "Del",
	})
	if err := svc.Delete(context.Background(), created.ID); err != nil {
		t.Fatal(err)
	}
}
