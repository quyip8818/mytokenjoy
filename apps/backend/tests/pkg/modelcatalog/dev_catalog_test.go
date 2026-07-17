package modelcatalog_test

import (
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
)

var (
	devModel1   = uuid.MustParse("00000000-0000-7000-0000-0000000000a1")
	devModel2   = uuid.MustParse("00000000-0000-7000-0000-0000000000a2")
	devModel100 = uuid.MustParse("00000000-0000-7000-0000-0000000000ba")
)

func TestModelLimitsCallTypes(t *testing.T) {
	t.Parallel()

	catalog := []types.ModelInfo{
		{ID: devModel1, Type: modelcatalog.DevCallTypeLocalTest, Enabled: true},
		{ID: devModel100, Type: "gpt-4o", Enabled: true},
		{ID: devModel2, Type: "disabled-dev", Enabled: false},
	}

	t.Run("local unions dev models", func(t *testing.T) {
		got := modelcatalog.ModelLimitsCallTypes(catalog, []uuid.UUID{devModel100}, true)
		if !slices.Contains(got, modelcatalog.DevCallTypeLocalTest) || !slices.Contains(got, "gpt-4o") {
			t.Fatalf("got %v", got)
		}
		if slices.Contains(got, "disabled-dev") {
			t.Fatalf("disabled dev leaked: %v", got)
		}
	})

	t.Run("non-local whitelist only", func(t *testing.T) {
		got := modelcatalog.ModelLimitsCallTypes(catalog, []uuid.UUID{devModel100}, false)
		if len(got) != 1 || got[0] != "gpt-4o" {
			t.Fatalf("got %v", got)
		}
	})
}

func TestIsLocalOnlyCallType(t *testing.T) {
	t.Parallel()
	if !modelcatalog.IsLocalOnlyCallType(modelcatalog.DevCallTypeLocalTest) {
		t.Fatal("expected dev-local-test to be local-only")
	}
	if modelcatalog.IsLocalOnlyCallType("gpt-4o") {
		t.Fatal("gpt-4o must not be local-only")
	}
}

func TestIsDevModel(t *testing.T) {
	t.Parallel()
	devModelInfo := types.ModelInfo{ID: devModel1, Type: "dev-local-test", Enabled: true}
	prodModelInfo := types.ModelInfo{ID: devModel100, Type: "gpt-4o", Enabled: true}
	if !modelcatalog.IsDevModel(devModelInfo) {
		t.Fatal("expected dev model")
	}
	if modelcatalog.IsDevModel(prodModelInfo) {
		t.Fatal("expected prod model not to be dev")
	}
}
