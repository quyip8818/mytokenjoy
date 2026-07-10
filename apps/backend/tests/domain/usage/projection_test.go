package usage_test

import (
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/domain/usage"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestProjectionUpsertBucketIdempotent(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()
	occurred := time.Date(2026, 6, 10, 9, 30, 0, 0, time.UTC)
	open := pkgbudget.OpenSnapshotKey(contract.DemoBudgetPeriod, clock.System())
	entry := types.UsageLedgerEntry{
		ID: "ledger-1", PlatformKeyID: "plk-1", DepartmentID: "dept-3",
		Model: "gpt-4o", Amount: 1.5, OccurredAt: occurred,
		PeriodKey:   contract.DemoBudgetPeriod,
		InputTokens: 100, OutputTokens: 50,
	}
	if err := usage.Apply(ctx, st, entry, open); err != nil {
		t.Fatal(err)
	}
	entry2 := entry
	entry2.ID = "ledger-2"
	entry2.Amount = 2.0
	if err := usage.Apply(ctx, st, entry2, open); err != nil {
		t.Fatal(err)
	}
	points, err := st.Usage().QuerySeries(ctx, types.UsageSeriesQuery{
		Granularity: types.UsageGranularityHour,
		Start:       occurred.Truncate(time.Hour),
		End:         occurred.Truncate(time.Hour).Add(time.Hour),
		GroupBy:     types.UsageGroupByNone,
		Timezone:    types.UsageDefaultTimezone,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(points) != 1 {
		t.Fatalf("expected one bucket, got %+v", points)
	}
	want := 3.5
	if points[0].Cost != want {
		t.Fatalf("expected aggregated cost %f, got %f", want, points[0].Cost)
	}
}

func TestProjectionUsesOrgPeriodKey(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()
	occurred := time.Date(2026, 7, 15, 9, 30, 0, 0, time.UTC)
	wantPeriod := contract.DemoBudgetPeriod
	calendarPeriod := pkgbudget.SnapshotKey(pkgbudget.PeriodMonthly, occurred)
	if calendarPeriod == wantPeriod {
		t.Fatal("test setup: calendar period must differ from org period")
	}

	beforeOrg := testutil.SnapshotConsumedAtPeriod(t, st, store.SnapshotAxisOrgNode, contract.IDDept3, wantPeriod)
	beforeCalendar := testutil.SnapshotConsumedAtPeriod(t, st, store.SnapshotAxisOrgNode, contract.IDDept3, calendarPeriod)

	open := pkgbudget.OpenSnapshotKey(wantPeriod, clock.System())
	entry := types.UsageLedgerEntry{
		ID: "ledger-period", PlatformKeyID: contract.IDPlatformKey1, DepartmentID: contract.IDDept3,
		Model: "gpt-4o", Amount: 1.0, OccurredAt: occurred,
		PeriodKey: wantPeriod,
	}
	if err := usage.Apply(ctx, st, entry, open); err != nil {
		t.Fatal(err)
	}

	afterOrg := testutil.SnapshotConsumedAtPeriod(t, st, store.SnapshotAxisOrgNode, contract.IDDept3, wantPeriod)
	if afterOrg != beforeOrg+1.0 {
		t.Fatalf("expected org period consumption +1, before=%v after=%v", beforeOrg, afterOrg)
	}
	afterCalendar := testutil.SnapshotConsumedAtPeriod(t, st, store.SnapshotAxisOrgNode, contract.IDDept3, calendarPeriod)
	if afterCalendar != beforeCalendar {
		t.Fatalf("expected no change at calendar period %q, before=%v after=%v", calendarPeriod, beforeCalendar, afterCalendar)
	}
}
