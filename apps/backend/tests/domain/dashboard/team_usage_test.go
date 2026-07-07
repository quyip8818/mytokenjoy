package dashboard_test

import (
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestTeamUsageTopModelBatch(t *testing.T) {
	t.Parallel()
	svc, st := newDashboardSvc(t)
	ctx := testutil.Ctx()
	base := time.Date(2026, 6, 19, 10, 0, 0, 0, time.UTC)
	testutil.SeedUsageBucket(t, st, testutil.UsageBucketOpts{
		DepartmentID: seed.IDDept3,
		Model:        "gpt-4o",
		CostCNY:      30,
		BucketStart:  base,
	})
	testutil.SeedUsageBucket(t, st, testutil.UsageBucketOpts{
		DepartmentID: seed.IDDept3,
		Model:        "gpt-4o-mini",
		CostCNY:      5,
		BucketStart:  base,
	})
	testutil.SeedUsageBucket(t, st, testutil.UsageBucketOpts{
		DepartmentID: seed.IDDept4,
		Model:        "claude-3",
		CostCNY:      20,
		BucketStart:  base,
	})

	teams, err := svc.TeamUsage(ctx, types.CostQueryParams{Period: string(types.CostPeriodCurrentMonth)}, testutil.AdminDashboardScope())
	if err != nil {
		t.Fatal(err)
	}
	var dept3Top, dept4Top string
	for _, team := range teams {
		switch team.DepartmentID {
		case seed.IDDept3:
			dept3Top = team.TopModel
		case seed.IDDept4:
			dept4Top = team.TopModel
		}
	}
	if dept3Top != "gpt-4o" {
		t.Fatalf("expected dept-3 top model gpt-4o, got %q", dept3Top)
	}
	if dept4Top != "claude-3" {
		t.Fatalf("expected dept-4 top model claude-3, got %q", dept4Top)
	}
}
