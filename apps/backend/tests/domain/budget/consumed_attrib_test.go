package budget_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestConsumptionDeltasMatchesApplyIncrement(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()
	open, err := pkgbudget.OpenDepartmentPeriod(ctx, st.Org().Nodes(), contract.IDDept3, cfg.Clock())
	if err != nil {
		t.Fatal(err)
	}

	memberID := contract.IDMember1
	groupID := contract.IDProject1
	entry := types.UsageLedgerEntry{
		PlatformKeyID: contract.IDPlatformKey1,
		DepartmentID:  contract.IDDept3,
		MemberID:      &memberID,
		ProjectID:     &groupID,
		Amount:        12.5,
	}

	deltas, err := budget.ConsumptionDeltas(ctx, st.Org().Nodes(), entry, open)
	if err != nil {
		t.Fatal(err)
	}
	if len(deltas) < 4 {
		t.Fatalf("expected at least 4 deltas, got %d", len(deltas))
	}

	periodKey := open.String()
	before := map[budget.AxisKey]float64{}
	for _, d := range deltas {
		key := budget.AxisKey{Kind: d.Kind, AxisID: d.AxisID, PeriodKey: d.PeriodKey}
		got, found, err := st.BudgetConsumed().GetConsumed(ctx, d.Kind, d.AxisID, periodKey)
		if err != nil {
			t.Fatal(err)
		}
		if !found {
			got = 0
		}
		before[key] = got
	}

	if err := budget.ApplyIncrement(ctx, st.BudgetConsumed(), st.Org().Nodes(), entry, open); err != nil {
		t.Fatal(err)
	}

	for _, d := range deltas {
		key := budget.AxisKey{Kind: d.Kind, AxisID: d.AxisID, PeriodKey: d.PeriodKey}
		got, found, err := st.BudgetConsumed().GetConsumed(ctx, d.Kind, d.AxisID, periodKey)
		if err != nil {
			t.Fatal(err)
		}
		if !found {
			t.Fatalf("expected consumed row for %s/%s", d.Kind, d.AxisID)
		}
		if got-before[key] != d.Amount {
			t.Fatalf("axis %s/%s expected +%v got +%v", d.Kind, d.AxisID, d.Amount, got-before[key])
		}
	}
}

func TestExpectedConsumedAggregatesMultipleEntries(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()

	entries := []types.UsageLedgerEntry{
		{PlatformKeyID: contract.IDPlatformKey1, DepartmentID: contract.IDDept3, Amount: 1},
		{PlatformKeyID: contract.IDPlatformKey1, DepartmentID: contract.IDDept3, Amount: 2},
	}
	expected, err := budget.ExpectedConsumed(ctx, st.Org().Nodes(), entries, cfg.Clock())
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for k, v := range expected {
		if k.Kind == store.AxisKindPlatformKey && k.AxisID == contract.IDPlatformKey1 && v == 3 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected platform key consumed 3 in map, got %#v", expected)
	}
}
