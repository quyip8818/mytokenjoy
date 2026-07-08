package postgres_test

import (
	"context"
	"testing"
	"time"

	orgfix "github.com/tokenjoy/backend/tests/testutil/org"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

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
	t.Parallel()
	st := testPostgresStore(t)
	ctx := testutil.Ctx()
	departments, err := common.LoadDepartments(ctx, st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	if len(departments) == 0 {
		t.Fatal("expected seeded departments")
	}
}

func TestRelayMappingRoundTrip(t *testing.T) {
	t.Parallel()
	st := testPostgresStore(t)
	ctx := testutil.Ctx()
	tokenID := int64(99001)
	memberID := contract.IDMember1
	mapping := store.RelayMapping{
		PlatformKeyID: contract.IDPlatformKey1,
		NewAPITokenID: &tokenID,
		MemberID:      &memberID,
		DepartmentID:  contract.IDDept3,
		SyncStatus:    store.RelaySyncStatusSynced,
		RelayGroup:    "dept-dept-3",
	}
	if err := st.Relay().UpsertMapping(ctx, mapping); err != nil {
		t.Fatal(err)
	}
	got, err := st.Relay().GetMappingByPlatformKeyID(ctx, contract.IDPlatformKey1)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected mapping round-trip")
	}
	if got.PlatformKeyID != mapping.PlatformKeyID {
		t.Fatalf("mapping mismatch: got %+v", got)
	}
	if got.DepartmentID != contract.IDDept3 {
		t.Fatalf("expected department from member join, got %q", got.DepartmentID)
	}
	if got.NewAPITokenID == nil || *got.NewAPITokenID != tokenID {
		t.Fatalf("expected token id %d, got %v", tokenID, got.NewAPITokenID)
	}
}

func TestMemberPersistAcrossRestart(t *testing.T) {
	t.Parallel()
	schemaURL := testutil.TestSchemaURL(t)
	ctx := testutil.Ctx()
	cfg := testutil.TestConfig()
	cfg.DatabaseURL = schemaURL

	st1, err := postgres.New(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	budgetTree, err := common.LoadBudgetTree(ctx, st1.Org().Nodes())
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
		if members[i].ID == contract.IDMember1 {
			members[i].Name = "PersistTest"
			updated = true
			break
		}
	}
	if !updated {
		t.Fatalf("member %s not found in seed", contract.IDMember1)
	}
	if err := st1.Org().SetMembers(ctx, members); err != nil {
		t.Fatal(err)
	}
	if pg, ok := st1.(*postgres.Store); ok {
		pg.Close()
	}

	st2 := reopenPostgresStore(t, schemaURL)
	members, err = st2.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got := findMemberName(members, contract.IDMember1); got != "PersistTest" {
		t.Fatalf("expected persisted member name, got %q", got)
	}
	budgetTree, err = common.LoadBudgetTree(ctx, st2.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	gotLen, gotRoot := budgetTreeSignature(budgetTree)
	if gotLen != budgetLen || gotRoot != budgetRoot {
		t.Fatalf("budget tree changed: before (%d,%q) after (%d,%q)", budgetLen, budgetRoot, gotLen, gotRoot)
	}
}

func TestWithTxCommitsDomainWrites(t *testing.T) {
	t.Parallel()
	st := testPostgresStore(t)
	pool := testDBPool(t)
	ctx := testutil.Ctx()

	modelsBefore := modelUpdatedAt(t, pool, "model-1")
	memberBefore := memberUpdatedAt(t, pool, contract.IDMember1)
	budgetBefore := budgetNodeUpdatedAt(t, pool, "dept-1")

	err := st.WithTx(ctx, func(tx store.Store) error {
		members, err := tx.Org().Members(ctx)
		if err != nil {
			return err
		}
		for i := range members {
			if members[i].ID == contract.IDMember1 {
				members[i].Name = "TxTest"
			}
		}
		if err := tx.Org().SetMembers(ctx, members); err != nil {
			return err
		}
		tree, err := common.LoadBudgetTree(ctx, tx.Org().Nodes())
		if err != nil {
			return err
		}
		if len(tree) > 0 {
			tree[0].Name = tree[0].Name + "-tx"
		}
		return orgfix.PersistBudgetTree(ctx, tx, tree)
	})
	if err != nil {
		t.Fatal(err)
	}

	memberAfter := memberUpdatedAt(t, pool, contract.IDMember1)
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
