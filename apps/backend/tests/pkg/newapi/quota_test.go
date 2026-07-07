package newapi_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
)

func TestToNewAPIUnitsUsesHighestWhitelistPrice(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	cny := newapi.CostCNYFromQuota(500000, 2)
	if cny != 2 {
		t.Fatalf("expected 2 CNY, got %v", cny)
	}
}

func TestHighestModelPriceCNY(t *testing.T) {
	t.Parallel()
	models := []types.ModelInfo{
		{Name: "cheap", InputPrice: 1, OutputPrice: 1},
		{Name: "mid", InputPrice: 5, OutputPrice: 5},
		{Name: "expensive", InputPrice: 10, OutputPrice: 10},
	}

	t.Run("all models", func(t *testing.T) {
		price := newapi.HighestModelPriceCNY(models, nil)
		if price != 20 {
			t.Errorf("expected 20, got %v", price)
		}
	})

	t.Run("whitelist subset", func(t *testing.T) {
		price := newapi.HighestModelPriceCNY(models, []string{"cheap", "mid"})
		if price != 10 {
			t.Errorf("expected 10, got %v", price)
		}
	})

	t.Run("empty models returns default", func(t *testing.T) {
		price := newapi.HighestModelPriceCNY(nil, nil)
		if price <= 0 {
			t.Errorf("expected positive default, got %v", price)
		}
	})
}

func TestFromNewAPIUnits(t *testing.T) {
	t.Parallel()
	models := []types.ModelInfo{
		{Name: "gpt-4", InputPrice: 10, OutputPrice: 10},
	}

	t.Run("positive units", func(t *testing.T) {
		cny := newapi.FromNewAPIUnits(500000, models, []string{"gpt-4"})
		if cny <= 0 {
			t.Errorf("expected positive CNY, got %v", cny)
		}
	})

	t.Run("zero units returns zero", func(t *testing.T) {
		cny := newapi.FromNewAPIUnits(0, models, nil)
		if cny != 0 {
			t.Errorf("expected 0, got %v", cny)
		}
	})

	t.Run("negative units returns zero", func(t *testing.T) {
		cny := newapi.FromNewAPIUnits(-100, models, nil)
		if cny != 0 {
			t.Errorf("expected 0, got %v", cny)
		}
	})
}

func TestToNewAPIUnitsZeroRemaining(t *testing.T) {
	t.Parallel()
	units := newapi.ToNewAPIUnits(0, nil, nil)
	if units != 0 {
		t.Errorf("expected 0, got %d", units)
	}
}

func TestFormatModelLimits(t *testing.T) {
	t.Parallel()
	t.Run("empty", func(t *testing.T) {
		result := newapi.FormatModelLimits(nil)
		if result != "" {
			t.Errorf("expected empty, got %q", result)
		}
	})

	t.Run("single", func(t *testing.T) {
		result := newapi.FormatModelLimits([]string{"gpt-4"})
		if result != "gpt-4" {
			t.Errorf("expected 'gpt-4', got %q", result)
		}
	})

	t.Run("multiple", func(t *testing.T) {
		result := newapi.FormatModelLimits([]string{"gpt-4", "gpt-3.5", "claude"})
		if result != "gpt-4,gpt-3.5,claude" {
			t.Errorf("expected 'gpt-4,gpt-3.5,claude', got %q", result)
		}
	})
}

func TestModelPriceCNY(t *testing.T) {
	t.Parallel()
	models := []types.ModelInfo{
		{Name: "gpt-4", InputPrice: 10, OutputPrice: 10},
		{Name: "free", InputPrice: 0, OutputPrice: 0},
	}

	t.Run("found model", func(t *testing.T) {
		price := newapi.ModelPriceCNY(models, "gpt-4")
		if price != 20 {
			t.Errorf("expected 20, got %v", price)
		}
	})

	t.Run("zero price returns default", func(t *testing.T) {
		price := newapi.ModelPriceCNY(models, "free")
		if price <= 0 {
			t.Errorf("expected positive default, got %v", price)
		}
	})

	t.Run("not found returns default", func(t *testing.T) {
		price := newapi.ModelPriceCNY(models, "unknown")
		if price <= 0 {
			t.Errorf("expected positive default, got %v", price)
		}
	})
}

func TestEffectiveWhitelist(t *testing.T) {
	t.Parallel()
	t.Run("empty key whitelist returns dept allowed", func(t *testing.T) {
		result := newapi.EffectiveWhitelist(nil, []string{"gpt-4", "claude"})
		if len(result) != 2 {
			t.Fatalf("expected 2, got %d", len(result))
		}
	})

	t.Run("intersection of key and dept", func(t *testing.T) {
		result := newapi.EffectiveWhitelist([]string{"gpt-4", "gpt-3.5"}, []string{"gpt-4", "claude"})
		if len(result) != 1 || result[0] != "gpt-4" {
			t.Errorf("expected [gpt-4], got %v", result)
		}
	})

	t.Run("no overlap returns empty", func(t *testing.T) {
		result := newapi.EffectiveWhitelist([]string{"gpt-3.5"}, []string{"gpt-4"})
		if len(result) != 0 {
			t.Errorf("expected empty, got %v", result)
		}
	})
}

func TestRelayGroupForDepartment(t *testing.T) {
	t.Parallel()
	group := newapi.RelayGroupForDepartment("dept-123")
	if group == "" {
		t.Error("expected non-empty group")
	}
	if group == "dept-123" {
		t.Error("expected group to have prefix")
	}
}
