//go:build testhook

package catalogsync_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/store"
)

func createTestCompany(t *testing.T, st store.Store) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	id := uuid.Must(uuid.NewV7())
	if err := st.Company().Create(ctx, store.Company{
		ID:   id,
		Name: "test-local-" + id.String()[:8],
	}); err != nil {
		t.Fatalf("createTestCompany: %v", err)
	}
	return id
}
