package dashboard_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/dashboard"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/tests/testutil"
)

func newDashboardService(t *testing.T) dashboard.Service {
	t.Helper()
	cfg, st := testutil.NewMemoryStoreFromConfig(t)
	return dashboard.NewService(cfg, st)
}

func TestCostSummaryByPeriod(t *testing.T) {
	svc := newDashboardService(t)
	current := svc.CostSummary(string(types.CostPeriodCurrentMonth))
	lastMonth := svc.CostSummary(string(types.CostPeriodLastMonth))
	if current.TotalCost <= lastMonth.TotalCost {
		t.Fatalf("expected current month cost > last month, got %v vs %v", current.TotalCost, lastMonth.TotalCost)
	}
}
