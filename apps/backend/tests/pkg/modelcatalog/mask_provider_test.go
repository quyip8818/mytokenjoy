package modelcatalog_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/support/modelcatalog"
)

func TestMaskProviderForTenantReplacesPlatformAndSeed(t *testing.T) {
	t.Parallel()
	models := []types.ModelInfo{
		{ID: uuid.New(), Provider: "openai", Source: "platform"},
		{ID: uuid.New(), Provider: "volcengine", Source: "seed"},
		{ID: uuid.New(), Provider: "custom", Source: "manual"},
	}
	modelcatalog.MaskProviderForTenant(models)
	if models[0].Provider != modelcatalog.DisplayProvider {
		t.Fatalf("platform model: expected provider %q, got %q", modelcatalog.DisplayProvider, models[0].Provider)
	}
	if models[1].Provider != modelcatalog.DisplayProvider {
		t.Fatalf("seed model: expected provider %q, got %q", modelcatalog.DisplayProvider, models[1].Provider)
	}
	if models[2].Provider != "custom" {
		t.Fatalf("manual model: expected provider %q, got %q", "custom", models[2].Provider)
	}
}

func TestMaskProviderForTenantEmptySlice(t *testing.T) {
	t.Parallel()
	// Should not panic on empty/nil input.
	modelcatalog.MaskProviderForTenant(nil)
	modelcatalog.MaskProviderForTenant([]types.ModelInfo{})
}
