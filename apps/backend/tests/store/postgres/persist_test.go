//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"

	"github.com/tokenjoy/backend/internal/seed"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/tests/testutil"
)

func testPostgresStore(t *testing.T) store.Store {
	t.Helper()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set")
	}
	cfg := testutil.TestConfig()
	st, err := postgres.New(context.Background(), dbURL, seed.Load(cfg))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if pg, ok := st.(*postgres.Store); ok {
			pg.Close()
		}
	})
	return st
}

func TestLoadOrSeedDomain(t *testing.T) {
	st := testPostgresStore(t)
	if len(st.Org().Departments()) == 0 {
		t.Fatal("expected seeded departments")
	}
}

func TestRelayMappingRoundTrip(t *testing.T) {
	st := testPostgresStore(t)
	tokenID := int64(99001)
	memberID := "m-1"
	mapping := store.RelayMapping{
		PlatformKeyID: "plk-persist-test",
		NewAPITokenID: &tokenID,
		MemberID:      &memberID,
		DepartmentID:  seed.IDDept3,
		SyncStatus:    store.RelaySyncStatusSynced,
		RelayGroup:    "dept-dept-3",
	}
	if err := st.Relay().UpsertMapping(mapping); err != nil {
		t.Fatal(err)
	}
	got, err := st.Relay().GetMappingByPlatformKeyID("plk-persist-test")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected mapping round-trip")
	}
	if got.PlatformKeyID != mapping.PlatformKeyID || got.DepartmentID != mapping.DepartmentID {
		t.Fatalf("mapping mismatch: got %+v", got)
	}
	if got.NewAPITokenID == nil || *got.NewAPITokenID != tokenID {
		t.Fatalf("expected token id %d, got %v", tokenID, got.NewAPITokenID)
	}
}
