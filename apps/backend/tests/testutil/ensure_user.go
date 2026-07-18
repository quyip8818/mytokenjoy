//go:build testhook

package testutil

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/store"
)

// EnsureUser inserts a user record if one does not already exist for the given ID.
// Tests that create ad-hoc members need a corresponding users row because the
// members query uses INNER JOIN users.
func EnsureUser(t *testing.T, st store.Store, userID uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	existing, err := st.User().GetByID(ctx, userID)
	if err != nil {
		t.Fatalf("EnsureUser: GetByID: %v", err)
	}
	if existing != nil {
		return
	}
	if err := st.User().Create(ctx, store.User{
		ID:     userID,
		Status: "active",
	}); err != nil {
		t.Fatalf("EnsureUser: Create: %v", err)
	}
}
