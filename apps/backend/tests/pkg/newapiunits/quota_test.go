package newapiunits_test

import (
	"math"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/newapiunits"
)

func TestToNewAPIUnitsUsesHighestWhitelistPrice(t *testing.T) {
	t.Parallel()
	models := []types.ModelInfo{
		{ModelID: 1, Type: "cheap", InputPrice: 1, OutputPrice: 1, Enabled: true},
		{ModelID: 2, Type: "expensive", InputPrice: 10, OutputPrice: 10, Enabled: true},
	}
	units := newapiunits.ToNewAPIUnits(20, models, []int64{1, 2})
	if units <= 0 {
		t.Fatalf("expected positive units, got %d", units)
	}
}

func TestCostFromQuota(t *testing.T) {
	t.Parallel()
	cost := newapiunits.CostFromQuota(500000, 2)
	if cost != 2 {
		t.Fatalf("expected 2, got %v", cost)
	}
}

func TestHighestModelPricePoint(t *testing.T) {
	t.Parallel()
	models := []types.ModelInfo{
		{ModelID: 1, Type: "cheap", InputPrice: 1, OutputPrice: 1, Enabled: true},
		{ModelID: 2, Type: "mid", InputPrice: 5, OutputPrice: 5, Enabled: true},
		{ModelID: 3, Type: "expensive", InputPrice: 10, OutputPrice: 10, Enabled: true},
	}

	t.Run("all models", func(t *testing.T) {
		price := newapiunits.HighestModelPricePoint(models, nil)
		if price != 20 {
			t.Errorf("expected 20, got %v", price)
		}
	})

	t.Run("whitelist subset", func(t *testing.T) {
		price := newapiunits.HighestModelPricePoint(models, []int64{1, 2})
		if price != 10 {
			t.Errorf("expected 10, got %v", price)
		}
	})

	t.Run("empty models returns default", func(t *testing.T) {
		price := newapiunits.HighestModelPricePoint(nil, nil)
		if price <= 0 {
			t.Errorf("expected positive default, got %v", price)
		}
	})
}

func TestFromNewAPIUnits(t *testing.T) {
	t.Parallel()
	models := []types.ModelInfo{
		{ModelID: 1, Type: "gpt-4", InputPrice: 10, OutputPrice: 10, Enabled: true},
	}

	t.Run("positive units", func(t *testing.T) {
		cny := newapiunits.FromNewAPIUnits(500000, models, []int64{1})
		if cny <= 0 {
			t.Errorf("expected positive CNY, got %v", cny)
		}
	})

	t.Run("zero units returns zero", func(t *testing.T) {
		cny := newapiunits.FromNewAPIUnits(0, models, nil)
		if cny != 0 {
			t.Errorf("expected 0, got %v", cny)
		}
	})

	t.Run("negative units returns zero", func(t *testing.T) {
		cny := newapiunits.FromNewAPIUnits(-100, models, nil)
		if cny != 0 {
			t.Errorf("expected 0, got %v", cny)
		}
	})
}

func TestToNewAPIUnitsZeroRemaining(t *testing.T) {
	t.Parallel()
	units := newapiunits.ToNewAPIUnits(0, nil, nil)
	if units != 0 {
		t.Errorf("expected 0, got %d", units)
	}
}

func TestToNewAPIUnitsSaturatesOverflow(t *testing.T) {
	t.Parallel()
	models := []types.ModelInfo{
		{ModelID: 1, Type: "tiny", InputPrice: 1e-12, OutputPrice: 0, Enabled: true},
	}
	units := newapiunits.ToNewAPIUnits(1e20, models, []int64{1})
	if units != math.MaxInt64 {
		t.Fatalf("expected MaxInt64, got %d", units)
	}
}

func TestQuotaDeltaClampsAddRoom(t *testing.T) {
	t.Parallel()
	current := int64(math.MaxInt64 - 10)
	delta := newapiunits.QuotaDelta(math.MaxInt64, current)
	if delta != 10 {
		t.Fatalf("expected delta 10, got %d", delta)
	}
	if current+delta != math.MaxInt64 {
		t.Fatalf("overflowing add: current+delta=%d", current+delta)
	}
	// Asking past MaxInt64 must not wrap.
	if got := newapiunits.QuotaDelta(math.MaxInt64, math.MaxInt64-5); got != 5 {
		t.Fatalf("expected 5, got %d", got)
	}
}

func TestQuotaDeltaSubtract(t *testing.T) {
	t.Parallel()
	got := newapiunits.QuotaDelta(100, 500)
	if got != -400 {
		t.Fatalf("expected -400, got %d", got)
	}
}

func TestQuotaDeltaNegativeCurrentTreatedAsZero(t *testing.T) {
	t.Parallel()
	got := newapiunits.QuotaDelta(10, -5)
	if got != 10 {
		t.Fatalf("expected 10, got %d", got)
	}
}

func TestAddSat(t *testing.T) {
	t.Parallel()
	if got := newapiunits.AddSat(math.MaxInt64-1, 5); got != math.MaxInt64 {
		t.Fatalf("expected MaxInt64, got %d", got)
	}
	if got := newapiunits.AddSat(10, 20); got != 30 {
		t.Fatalf("expected 30, got %d", got)
	}
}

func TestSubFloor0(t *testing.T) {
	t.Parallel()
	if got := newapiunits.SubFloor0(10, 3); got != 7 {
		t.Fatalf("expected 7, got %d", got)
	}
	if got := newapiunits.SubFloor0(3, 10); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
	if got := newapiunits.SubFloor0(math.MaxInt64, math.MaxInt64); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestFormatModelLimits(t *testing.T) {
	t.Parallel()
	t.Run("empty", func(t *testing.T) {
		result := newapiunits.FormatModelLimits(nil)
		if result != "" {
			t.Errorf("expected empty, got %q", result)
		}
	})

	t.Run("single", func(t *testing.T) {
		result := newapiunits.FormatModelLimits([]string{"gpt-4"})
		if result != "gpt-4" {
			t.Errorf("expected 'gpt-4', got %q", result)
		}
	})

	t.Run("multiple", func(t *testing.T) {
		result := newapiunits.FormatModelLimits([]string{"gpt-4", "gpt-3.5", "claude"})
		if result != "gpt-4,gpt-3.5,claude" {
			t.Errorf("expected 'gpt-4,gpt-3.5,claude', got %q", result)
		}
	})
}

func TestModelPricePoint(t *testing.T) {
	t.Parallel()
	models := []types.ModelInfo{
		{ModelID: 1, Type: "gpt-4", InputPrice: 10, OutputPrice: 10, Enabled: true},
		{ModelID: 2, Type: "free", InputPrice: 0, OutputPrice: 0, Enabled: true},
	}

	t.Run("found model", func(t *testing.T) {
		price := newapiunits.ModelPricePoint(models, []int64{1}, "gpt-4")
		if price != 20 {
			t.Errorf("expected 20, got %v", price)
		}
	})

	t.Run("zero price returns default", func(t *testing.T) {
		price := newapiunits.ModelPricePoint(models, []int64{2}, "free")
		if price <= 0 {
			t.Errorf("expected positive default, got %v", price)
		}
	})

	t.Run("not found returns default", func(t *testing.T) {
		price := newapiunits.ModelPricePoint(models, nil, "unknown")
		if price <= 0 {
			t.Errorf("expected positive default, got %v", price)
		}
	})
}

func TestEffectiveWhitelistIDs(t *testing.T) {
	t.Parallel()
	t.Run("empty key whitelist returns dept allowed", func(t *testing.T) {
		result := newapiunits.EffectiveWhitelistIDs(nil, []int64{1, 2})
		if len(result) != 2 {
			t.Fatalf("expected 2, got %d", len(result))
		}
	})

	t.Run("intersection of key and dept", func(t *testing.T) {
		result := newapiunits.EffectiveWhitelistIDs([]int64{1, 3}, []int64{1, 2})
		if len(result) != 1 || result[0] != 1 {
			t.Errorf("expected [1], got %v", result)
		}
	})

	t.Run("no overlap returns empty", func(t *testing.T) {
		result := newapiunits.EffectiveWhitelistIDs([]int64{3}, []int64{1})
		if len(result) != 0 {
			t.Errorf("expected empty, got %v", result)
		}
	})
}

func TestNewAPIGroupForDepartment(t *testing.T) {
	t.Parallel()
	group := newapiunits.NewAPIGroupForDepartment("dept-123")
	if group == "" {
		t.Error("expected non-empty group")
	}
	if group == "dept-123" {
		t.Error("expected group to have prefix")
	}
}
