//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/internal/store/seed"
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

func memberUpdatedAt(t *testing.T, pool *pgxpool.Pool, memberID string) time.Time {
	t.Helper()
	var ts time.Time
	err := pool.QueryRow(context.Background(), `
		SELECT updated_at FROM members WHERE id = $1
	`, memberID).Scan(&ts)
	if err != nil {
		t.Fatalf("read member %s updated_at: %v", memberID, err)
	}
	return ts
}

func budgetNodeUpdatedAt(t *testing.T, pool *pgxpool.Pool, nodeID string) time.Time {
	t.Helper()
	var ts time.Time
	err := pool.QueryRow(context.Background(), `
		SELECT updated_at FROM org_nodes WHERE id = $1
	`, nodeID).Scan(&ts)
	if err != nil {
		t.Fatalf("read budget node %s updated_at: %v", nodeID, err)
	}
	return ts
}

func modelUpdatedAt(t *testing.T, pool *pgxpool.Pool, modelID string) time.Time {
	t.Helper()
	var ts time.Time
	err := pool.QueryRow(context.Background(), `
		SELECT updated_at FROM models WHERE id = $1
	`, modelID).Scan(&ts)
	if err != nil {
		t.Fatalf("read model %s updated_at: %v", modelID, err)
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
	ctx := testutil.Ctx()
	departments, err := common.LoadDepartments(ctx, st)
	if err != nil {
		t.Fatal(err)
	}
	if len(departments) == 0 {
		t.Fatal("expected seeded departments")
	}
}

func TestRelayMappingRoundTrip(t *testing.T) {
	st := testPostgresStore(t)
	ctx := testutil.Ctx()
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
	if err := st.Relay().UpsertMapping(ctx, mapping); err != nil {
		t.Fatal(err)
	}
	got, err := st.Relay().GetMappingByPlatformKeyID(ctx, "plk-persist-test")
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

func TestMemberPersistAcrossRestart(t *testing.T) {
	dbURL := requireDatabaseURL(t)
	ctx := testutil.Ctx()
	cfg := testutil.TestConfig()
	cfg.DatabaseURL = dbURL

	st1, err := postgres.New(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	budgetTree, err := common.LoadBudgetTree(ctx, st1)
	if err != nil {
		t.Fatal(err)
	}
	budgetLen, budgetRoot := budgetTreeSignature(budgetTree)
	members, err := st1.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	updated := false
	for i := range members {
		if members[i].ID == seed.IDMember1 {
			members[i].Name = "PersistTest"
			updated = true
			break
		}
	}
	if !updated {
		t.Fatalf("member %s not found in seed", seed.IDMember1)
	}
	if err := st1.Org().SetMembers(ctx, members); err != nil {
		t.Fatal(err)
	}
	if pg, ok := st1.(*postgres.Store); ok {
		pg.Close()
	}

	st2 := reopenPostgresStore(t, dbURL)
	members, err = st2.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got := findMemberName(members, seed.IDMember1); got != "PersistTest" {
		t.Fatalf("expected persisted member name, got %q", got)
	}
	budgetTree, err = common.LoadBudgetTree(ctx, st2)
	if err != nil {
		t.Fatal(err)
	}
	gotLen, gotRoot := budgetTreeSignature(budgetTree)
	if gotLen != budgetLen || gotRoot != budgetRoot {
		t.Fatalf("budget tree changed: before (%d,%q) after (%d,%q)", budgetLen, budgetRoot, gotLen, gotRoot)
	}
}

func TestWithTxCommitsDomainWrites(t *testing.T) {
	st := testPostgresStore(t)
	pool := testDBPool(t)
	ctx := testutil.Ctx()

	modelsBefore := modelUpdatedAt(t, pool, "model-1")
	memberBefore := memberUpdatedAt(t, pool, seed.IDMember1)
	budgetBefore := budgetNodeUpdatedAt(t, pool, "dept-1")

	err := st.WithTx(ctx, func(tx store.Store) error {
		members, err := tx.Org().Members(ctx)
		if err != nil {
			return err
		}
		for i := range members {
			if members[i].ID == seed.IDMember1 {
				members[i].Name = "TxTest"
			}
		}
		if err := tx.Org().SetMembers(ctx, members); err != nil {
			return err
		}
		tree, err := common.LoadBudgetTree(ctx, tx)
		if err != nil {
			return err
		}
		if len(tree) > 0 {
			tree[0].Name = tree[0].Name + "-tx"
		}
		return testutil.PersistBudgetTree(ctx, tx, tree)
	})
	if err != nil {
		t.Fatal(err)
	}

	memberAfter := memberUpdatedAt(t, pool, seed.IDMember1)
	budgetAfter := budgetNodeUpdatedAt(t, pool, "dept-1")
	modelsAfter := modelUpdatedAt(t, pool, "model-1")

	if !memberAfter.After(memberBefore) {
		t.Fatalf("expected member updated_at to advance: before=%v after=%v", memberBefore, memberAfter)
	}
	if !budgetAfter.After(budgetBefore) {
		t.Fatalf("expected budget node updated_at to advance: before=%v after=%v", budgetBefore, budgetAfter)
	}
	if !modelsAfter.Equal(modelsBefore) {
		t.Fatalf("expected model updated_at unchanged: before=%v after=%v", modelsBefore, modelsAfter)
	}
}
