package newapi_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
)

func TestToNewAPIUnitsUsesHighestWhitelistPrice(t *testing.T) {
	models := []types.ModelInfo{
		{Name: "cheap", InputPrice: 1, OutputPrice: 1},
		{Name: "expensive", InputPrice: 10, OutputPrice: 10},
	}
	units := newapi.ToNewAPIUnits(20, models, []string{"cheap", "expensive"})
	if units <= 0 {
		t.Fatalf("expected positive units, got %d", units)
	}
}

func TestCostCNYFromQuota(t *testing.T) {
	cny := newapi.CostCNYFromQuota(500000, 2)
	if cny != 2 {
		t.Fatalf("expected 2 CNY, got %v", cny)
	}
}
