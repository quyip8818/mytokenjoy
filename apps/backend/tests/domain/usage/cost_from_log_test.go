package usage_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/store"
)

func TestCostCNYFromLog(t *testing.T) {
	t.Parallel()
	models := []types.ModelInfo{{Type: "gpt-4o", InputPrice: 1, OutputPrice: 1}}
	cost := domainusage.CostCNYFromLog(500000, "gpt-4o", models, nil)
	if cost != 2 {
		t.Fatalf("expected cost 2, got %v", cost)
	}
}

func TestResolveConsumeModel(t *testing.T) {
	t.Parallel()
	model := domainusage.ResolveConsumeModel(store.RawConsumeLog{ModelName: "gpt-4o"})
	if model != "gpt-4o" {
		t.Fatalf("expected gpt-4o, got %s", model)
	}
}
