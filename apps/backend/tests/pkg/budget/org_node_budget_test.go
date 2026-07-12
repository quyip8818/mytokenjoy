package budget_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestOrgNodeBudgetRowFromNode_DefaultPeriod(t *testing.T) {
	t.Parallel()
	row := pkgbudget.OrgNodeBudgetRowFromNode(types.OrgNode{
		ID: "dept-a", Budget: 100, Period: "",
	})
	if row.Period != pkgbudget.PeriodMonthly {
		t.Fatalf("expected default period %q, got %q", pkgbudget.PeriodMonthly, row.Period)
	}
	if row.NodeID != "dept-a" || row.Budget != 100 {
		t.Fatalf("unexpected row: %+v", row)
	}
}

func TestOrgNodeBudgetRowFromNode_PreservesFields(t *testing.T) {
	t.Parallel()
	reserved := floatPtr(500)
	row := pkgbudget.OrgNodeBudgetRowFromNode(types.OrgNode{
		ID: "dept-b", Budget: 2000, ReservedPool: reserved,
		Period: "2026-06", MemberAvgBudget: 300,
	})
	if row.ReservedPool == nil || *row.ReservedPool != 500 {
		t.Fatalf("expected reserved pool 500, got %+v", row.ReservedPool)
	}
	if row.Period != "2026-06" || row.MemberAvgBudget != 300 {
		t.Fatalf("unexpected row: %+v", row)
	}
}

func TestOrgNodeBudgetRowsFromNodes_FlattensTree(t *testing.T) {
	t.Parallel()
	nodes := []types.OrgNode{
		{
			ID: "root", Budget: 1000, Period: pkgbudget.PeriodMonthly,
			Children: []types.OrgNode{
				{ID: "child-a", Budget: 400},
				{ID: "child-b", Budget: 600, ReservedPool: floatPtr(50)},
			},
		},
	}
	rows := pkgbudget.OrgNodeBudgetRowsFromNodes(nodes)
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}
	byID := make(map[string]store.OrgNodeBudgetRow, len(rows))
	for _, row := range rows {
		byID[row.NodeID] = row
	}
	if byID["root"].Budget != 1000 {
		t.Fatalf("root budget mismatch: %+v", byID["root"])
	}
	if byID["child-b"].ReservedPool == nil || *byID["child-b"].ReservedPool != 50 {
		t.Fatalf("child-b reserved mismatch: %+v", byID["child-b"])
	}
}

func TestOrgNodeBudgetRowsFromNodes_MatchesSnapshotDept3(t *testing.T) {
	t.Parallel()
	wantBudget := testutil.DisplayPoints(20000)
	wantReserved := testutil.DisplayPoints(1500)
	rows := pkgbudget.OrgNodeBudgetRowsFromNodes([]types.OrgNode{
		{
			ID: "dept-3", Budget: wantBudget, ReservedPool: &wantReserved,
			Period: pkgbudget.PeriodMonthly,
		},
	})
	if len(rows) != 1 {
		t.Fatal("expected single row")
	}
	row := rows[0]
	if row.Budget != wantBudget {
		t.Fatalf("budget mismatch: got %v want %v", row.Budget, wantBudget)
	}
	if row.ReservedPool == nil || *row.ReservedPool != wantReserved {
		t.Fatalf("reserved mismatch: got %+v want %v", row.ReservedPool, wantReserved)
	}
}
