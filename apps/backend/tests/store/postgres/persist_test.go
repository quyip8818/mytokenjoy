//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/tests/testutil"
)

func requireDatabaseURL(t *testing.T) string {
	t.Helper()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set")
	}
	return dbURL
}

func testDBPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), requireDatabaseURL(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pool.Close() })
	return pool
}

func shardUpdatedAt(t *testing.T, pool *pgxpool.Pool, shard string) time.Time {
	t.Helper()
	var ts time.Time
	err := pool.QueryRow(context.Background(), `
		SELECT updated_at FROM domain_snapshot WHERE id = $1
	`, shard).Scan(&ts)
	if err != nil {
		t.Fatalf("read shard %s updated_at: %v", shard, err)
	}
	return ts
}

func reopenPostgresStore(t *testing.T, dbURL string) store.Store {
	t.Helper()
	cfg := testutil.TestConfig()
	cfg.DatabaseURL = dbURL
	st, err := postgres.New(context.Background(), cfg)
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

func testPostgresStore(t *testing.T) store.Store {
	t.Helper()
	return reopenPostgresStore(t, requireDatabaseURL(t))
}

func findMemberName(members []types.Member, id string) string {
	for _, member := range members {
		if member.ID == id {
			return member.Name
		}
	}
	return ""
}

func budgetTreeSignature(tree []types.BudgetNode) (int, string) {
	if len(tree) == 0 {
		return 0, ""
	}
	return len(tree), tree[0].Name
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

func TestShardPersistOrgOnly(t *testing.T) {
	dbURL := requireDatabaseURL(t)
	ctx := context.Background()
	cfg := testutil.TestConfig()
	cfg.DatabaseURL = dbURL

	st1, err := postgres.New(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	budgetLen, budgetRoot := budgetTreeSignature(st1.Budget().Tree())
	members := st1.Org().Members()
	updated := false
	for i := range members {
		if members[i].ID == seed.IDMember1 {
			members[i].Name = "ShardPersistTest"
			updated = true
			break
		}
	}
	if !updated {
		t.Fatalf("member %s not found in seed", seed.IDMember1)
	}
	if err := st1.Org().SetMembers(members); err != nil {
		t.Fatal(err)
	}
	if pg, ok := st1.(*postgres.Store); ok {
		pg.Close()
	}

	st2 := reopenPostgresStore(t, dbURL)
	if got := findMemberName(st2.Org().Members(), seed.IDMember1); got != "ShardPersistTest" {
		t.Fatalf("expected persisted member name, got %q", got)
	}
	gotLen, gotRoot := budgetTreeSignature(st2.Budget().Tree())
	if gotLen != budgetLen || gotRoot != budgetRoot {
		t.Fatalf("budget tree changed: before (%d,%q) after (%d,%q)", budgetLen, budgetRoot, gotLen, gotRoot)
	}
}

func TestWithTxFlushesDirtyShards(t *testing.T) {
	st := testPostgresStore(t)
	pool := testDBPool(t)
	ctx := context.Background()

	modelsBefore := shardUpdatedAt(t, pool, store.ShardModels)
	orgBefore := shardUpdatedAt(t, pool, store.ShardOrg)
	budgetBefore := shardUpdatedAt(t, pool, store.ShardBudget)

	err := st.WithTx(ctx, func(tx store.Store) error {
		members := tx.Org().Members()
		for i := range members {
			if members[i].ID == seed.IDMember1 {
				members[i].Name = "TxShardTest"
			}
		}
		if err := tx.Org().SetMembers(members); err != nil {
			return err
		}
		tree := tx.Budget().Tree()
		if len(tree) > 0 {
			tree[0].Name = tree[0].Name + "-tx"
		}
		return tx.Budget().SetTree(tree)
	})
	if err != nil {
		t.Fatal(err)
	}

	modelsAfter := shardUpdatedAt(t, pool, store.ShardModels)
	orgAfter := shardUpdatedAt(t, pool, store.ShardOrg)
	budgetAfter := shardUpdatedAt(t, pool, store.ShardBudget)

	if !orgAfter.After(orgBefore) {
		t.Fatalf("expected org shard updated_at to advance: before=%v after=%v", orgBefore, orgAfter)
	}
	if !budgetAfter.After(budgetBefore) {
		t.Fatalf("expected budget shard updated_at to advance: before=%v after=%v", budgetBefore, budgetAfter)
	}
	if !modelsAfter.Equal(modelsBefore) {
		t.Fatalf("expected models shard updated_at unchanged: before=%v after=%v", modelsBefore, modelsAfter)
	}
}
