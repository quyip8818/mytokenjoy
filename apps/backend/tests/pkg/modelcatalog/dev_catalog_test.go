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
		{ID: devModel1, Type: modelcatalog.TestCallType, Enabled: true},
		{ID: devModel100, Type: "gpt-4o", Enabled: true},
		{ID: devModel2, Type: "disabled-dev", Enabled: false},
	}

	t.Run("local unions test models", func(t *testing.T) {
		got := modelcatalog.ModelLimitsCallTypes(catalog, []uuid.UUID{devModel100}, true)
		if !slices.Contains(got, modelcatalog.TestCallType) || !slices.Contains(got, "gpt-4o") {
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

func TestIsTestOnlyCallType(t *testing.T) {
	t.Parallel()
	if !modelcatalog.IsTestOnlyCallType(modelcatalog.TestCallType) {
		t.Fatal("expected test-model to be test-only")
	}
	if modelcatalog.IsTestOnlyCallType("gpt-4o") {
		t.Fatal("gpt-4o must not be test-only")
	}
}

func TestIsTestModel(t *testing.T) {
	t.Parallel()
	testModelInfo := types.ModelInfo{ID: devModel1, Type: "test-model", Enabled: true}
	prodModelInfo := types.ModelInfo{ID: devModel100, Type: "gpt-4o", Enabled: true}
	if !modelcatalog.IsTestModel(testModelInfo) {
		t.Fatal("expected test model")
	}
	if modelcatalog.IsTestModel(prodModelInfo) {
		t.Fatal("expected prod model not to be test")
	}
}
