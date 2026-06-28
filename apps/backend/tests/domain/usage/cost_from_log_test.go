package usage_test

import (
	"testing"

	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
)

func TestCostCNYFromLog(t *testing.T) {
	models := []types.ModelInfo{{Name: "gpt-4o", InputPrice: 1, OutputPrice: 1}}
	cost := domainusage.CostCNYFromLog(500000, "gpt-4o", models)
	if cost != 2 {
		t.Fatalf("expected cost 2, got %v", cost)
	}
}

func TestResolveWebhookModel(t *testing.T) {
	model := domainusage.ResolveWebhookModel(newapi.WebhookLogPayload{Model: "gpt-4o"})
	if model != "gpt-4o" {
		t.Fatalf("expected gpt-4o, got %s", model)
	}
}
