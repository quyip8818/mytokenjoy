package modelcatalog_test

import (
	"slices"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
)

func TestModelLimitsCallTypes(t *testing.T) {
	t.Parallel()

	catalog := []types.ModelInfo{
		{ModelID: 1, Type: modelcatalog.DevCallTypeLocalTest, Enabled: true},
		{ModelID: 100, Type: "gpt-4o", Enabled: true},
		{ModelID: 2, Type: "disabled-dev", Enabled: false},
	}

	t.Run("local unions id<100", func(t *testing.T) {
		got := modelcatalog.ModelLimitsCallTypes(catalog, []int64{100}, true)
		if !slices.Contains(got, modelcatalog.DevCallTypeLocalTest) || !slices.Contains(got, "gpt-4o") {
			t.Fatalf("got %v", got)
		}
		if slices.Contains(got, "disabled-dev") {
			t.Fatalf("disabled dev leaked: %v", got)
		}
	})

	t.Run("non-local whitelist only", func(t *testing.T) {
		got := modelcatalog.ModelLimitsCallTypes(catalog, []int64{100}, false)
		if len(got) != 1 || got[0] != "gpt-4o" {
			t.Fatalf("got %v", got)
		}
	})
}

func TestIsLocalOnlyCallType(t *testing.T) {
	t.Parallel()
	if !modelcatalog.IsLocalOnlyCallType(modelcatalog.DevCallTypeLocalTest) {
		t.Fatal("expected local-test-model")
	}
	if modelcatalog.IsLocalOnlyCallType("gpt-4o") {
		t.Fatal("gpt-4o must not be local-only")
	}
	if !modelcatalog.IsDevCatalogModelID(1) || modelcatalog.IsDevCatalogModelID(100) {
		t.Fatal("ProdCatalogModelIDStart bounds")
	}
}
