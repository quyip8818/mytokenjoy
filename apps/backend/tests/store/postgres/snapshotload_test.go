package postgres_test

import (
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestLoadPlatformKeysWithUsedResolvesDepartmentPeriod(t *testing.T) {
	t.Parallel()
	st := testPostgresStore(t)
	ctx := testutil.Ctx()

	nodes, err := st.Org().Nodes().Tree(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !setOrgNodePeriod(nodes, contract.IDDept3, "2026-07") {
		t.Fatal("dept-3 not found")
	}
	if err := st.Org().Nodes().SetTree(ctx, nodes); err != nil {
		t.Fatal(err)
	}

	periodJune := pkgbudget.SnapshotKey("2026-06", time.Now().UTC())
	periodJuly := pkgbudget.SnapshotKey("2026-07", time.Now().UTC())
	testutil.SetSnapshotConsumedAtPeriod(t, st, store.SnapshotAxisPlatformKey, contract.IDPlatformKey1, periodJune, 99)
	testutil.SetSnapshotConsumedAtPeriod(t, st, store.SnapshotAxisPlatformKey, contract.IDPlatformKey1, periodJuly, 42)

	keys, err := pkgbudget.LoadPlatformKeysWithUsed(ctx, st.BudgetSnapshots(), st.Org(), st.Budget(), st.Keys())
	if err != nil {
		t.Fatal(err)
	}
	var used float64
	for _, key := range keys {
		if key.ID == contract.IDPlatformKey1 {
			used = key.Used
			break
		}
	}
	if used != 42 {
		t.Fatalf("expected plk-1 used=42 from dept-3 period 2026-07, got %v", used)
	}
}

func TestLoadBudgetGroupsWithConsumedSumsAcrossDepartmentPeriods(t *testing.T) {
	t.Parallel()
	st := testPostgresStore(t)
	ctx := testutil.Ctx()

	nodes, err := st.Org().Nodes().Tree(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !setOrgNodePeriod(nodes, contract.IDDept3, "2026-06") {
		t.Fatal("dept-3 not found")
	}
	if !setOrgNodePeriod(nodes, contract.IDDept4, "2026-07") {
		t.Fatal("dept-4 not found")
	}
	if err := st.Org().Nodes().SetTree(ctx, nodes); err != nil {
		t.Fatal(err)
	}

	periodJune := pkgbudget.SnapshotKey("2026-06", time.Now().UTC())
	periodJuly := pkgbudget.SnapshotKey("2026-07", time.Now().UTC())
	testutil.SetSnapshotConsumedAtPeriod(t, st, store.SnapshotAxisBudgetGroup, contract.IDBudgetGroup1, periodJune, 10)
	testutil.SetSnapshotConsumedAtPeriod(t, st, store.SnapshotAxisBudgetGroup, contract.IDBudgetGroup1, periodJuly, 7)

	groups, err := pkgbudget.LoadBudgetGroupsWithConsumed(ctx, st.BudgetSnapshots(), st.Org(), st.Budget())
	if err != nil {
		t.Fatal(err)
	}
	var consumed float64
	for _, group := range groups {
		if group.ID == contract.IDBudgetGroup1 {
			consumed = group.Consumed
			break
		}
	}
	if consumed != 17 {
		t.Fatalf("expected bg-1 consumed=17, got %v", consumed)
	}
}

func setOrgNodePeriod(nodes []types.OrgNode, id, period string) bool {
	for i := range nodes {
		if nodes[i].ID == id {
			nodes[i].Period = period
			return true
		}
		if len(nodes[i].Children) > 0 && setOrgNodePeriod(nodes[i].Children, id, period) {
			return true
		}
	}
	return false
}
