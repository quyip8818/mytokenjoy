package postgres_test

import (
	"testing"
	"time"

	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	budgetfix "github.com/tokenjoy/backend/tests/testutil/budget"
)

func TestLoadPlatformKeysWithUsedResolvesDepartmentPeriod(t *testing.T) {
	t.Parallel()
	st := testPostgresStore(t)
	ctx := testutil.Ctx()

	clk := clock.Fixed(time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC))
	budgetfix.SetSnapshotConsumedAtPeriod(t, st, store.AxisKindPlatformKey, contract.IDPlatformKey1, "2026-06", 99)
	budgetfix.SetSnapshotConsumedAtPeriod(t, st, store.AxisKindPlatformKey, contract.IDPlatformKey1, "2026-07", 42)

	keys, err := pkgbudget.LoadPlatformKeysWithUsed(ctx, st.BudgetConsumed(), st.Org(), st.Budget(), st.Keys(), clk)
	if err != nil {
		t.Fatal(err)
	}
	var used float64
	for _, key := range keys {
		if key.ID == contract.IDPlatformKey1 {
			used = key.Consumed
			break
		}
	}
	if used != 42 {
		t.Fatalf("expected plk-1 used=42 from open period 2026-07, got %v", used)
	}
}

func TestLoadProjectsWithConsumedUsesOpenPeriod(t *testing.T) {
	t.Parallel()
	st := testPostgresStore(t)
	ctx := testutil.Ctx()

	clk := clock.Fixed(time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC))
	budgetfix.SetSnapshotConsumedAtPeriod(t, st, store.AxisKindProject, contract.IDProject1, "2026-06", 10)
	budgetfix.SetSnapshotConsumedAtPeriod(t, st, store.AxisKindProject, contract.IDProject1, "2026-07", 7)

	projects, err := pkgbudget.LoadProjectsWithConsumed(ctx, st.BudgetConsumed(), st.Org(), st.Budget(), clk)
	if err != nil {
		t.Fatal(err)
	}
	var consumed float64
	for _, project := range projects {
		if project.ID == contract.IDProject1 {
			consumed = project.Consumed
			break
		}
	}
	if consumed != 7 {
		t.Fatalf("expected proj-1 consumed=7 from open period 2026-07, got %v", consumed)
	}
}
