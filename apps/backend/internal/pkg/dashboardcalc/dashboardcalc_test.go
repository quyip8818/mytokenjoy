package dashboardcalc_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/dashboardcalc"
)

func TestBuildCostSummaryScalesByPeriod(t *testing.T) {
	members := []types.Member{
		{ID: "m-1", Status: "active"},
		{ID: "m-2", Status: "active"},
	}
	current := dashboardcalc.BuildCostSummary(types.CostPeriodCurrentMonth, members)
	lastMonth := dashboardcalc.BuildCostSummary(types.CostPeriodLastMonth, members)
	if current.TotalCost <= lastMonth.TotalCost {
		t.Fatalf("expected current month cost > last month, got %v vs %v", current.TotalCost, lastMonth.TotalCost)
	}
}

func TestGetDepartmentCostsForParentPercentages(t *testing.T) {
	rows := dashboardcalc.GetDepartmentCostsForParent("", types.CostPeriodCurrentMonth)
	if len(rows) != 4 {
		t.Fatalf("expected 4 top-level departments, got %d", len(rows))
	}
	totalPct := 0.0
	for _, row := range rows {
		totalPct += row.Percentage
	}
	if totalPct < 99 || totalPct > 101 {
		t.Fatalf("expected percentages near 100, got %v", totalPct)
	}
}
