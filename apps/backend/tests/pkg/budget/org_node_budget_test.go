package budget_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
)

func TestOrgNodeBudgetRowFromNode_DefaultPeriod(t *testing.T) {
	t.Parallel()
	deptA := uuid.MustParse("00000000-0000-7000-0000-00000000da01")
	row := pkgbudget.OrgNodeBudgetRowFromNode(types.OrgNode{
		ID: deptA, Budget: 100, Period: "",
	})
	if row.Period != pkgbudget.PeriodMonthly {
		t.Fatalf("expected default period %q, got %q", pkgbudget.PeriodMonthly, row.Period)
	}
	if row.NodeID != deptA || row.Budget != 100 {
		t.Fatalf("unexpected row: %+v", row)
	}
}

func TestOrgNodeBudgetRowFromNode_PreservesFields(t *testing.T) {
	t.Parallel()
	deptB := uuid.MustParse("00000000-0000-7000-0000-00000000da02")
	reserved := budgetfix.FloatPtr(500)
	row := pkgbudget.OrgNodeBudgetRowFromNode(types.OrgNode{
		ID: deptB, Budget: 2000, ReservedPool: reserved,
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
	rootID := uuid.MustParse("00000000-0000-7000-0000-00000000da10")
	childA := uuid.MustParse("00000000-0000-7000-0000-00000000da11")
	childB := uuid.MustParse("00000000-0000-7000-0000-00000000da12")
	nodes := []types.OrgNode{
		{
			ID: rootID, Budget: 1000, Period: pkgbudget.PeriodMonthly,
			Children: []types.OrgNode{
				{ID: childA, Budget: 400},
				{ID: childB, Budget: 600, ReservedPool: budgetfix.FloatPtr(50)},
			},
		},
	}
	rows := pkgbudget.OrgNodeBudgetRowsFromNodes(nodes)
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}
	byID := make(map[uuid.UUID]store.OrgNodeBudgetRow, len(rows))
	for _, row := range rows {
		byID[row.NodeID] = row
	}
	if byID[rootID].Budget != 1000 {
		t.Fatalf("root budget mismatch: %+v", byID[rootID])
	}
	if byID[childB].ReservedPool == nil || *byID[childB].ReservedPool != 50 {
		t.Fatalf("child-b reserved mismatch: %+v", byID[childB])
	}
}

func TestOrgNodeBudgetRowsFromNodes_MatchesSnapshotDept3(t *testing.T) {
	t.Parallel()
	dept3 := uuid.MustParse("00000000-0000-7000-0000-00000000da03")
	wantBudget := budgetfix.DisplayPoints(20000)
	wantReserved := budgetfix.DisplayPoints(1500)
	rows := pkgbudget.OrgNodeBudgetRowsFromNodes([]types.OrgNode{
		{
			ID: dept3, Budget: wantBudget, ReservedPool: &wantReserved,
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
