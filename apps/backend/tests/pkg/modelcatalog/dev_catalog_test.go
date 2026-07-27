package modelcatalog_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
)

var (
	devModel1   = uuid.MustParse("00000000-0000-7000-0000-0000000000a1")
	devModel100 = uuid.MustParse("00000000-0000-7000-0000-0000000000ba")
)

func TestIsTestModel(t *testing.T) {
	t.Parallel()
	testModelInfo := types.ModelInfo{ID: devModel1, Type: "test-model", Source: modelcatalog.SourceTest, Enabled: true}
	prodModelInfo := types.ModelInfo{ID: devModel100, Type: "gpt-4o", Source: "sms", Enabled: true}
	if !modelcatalog.IsTestModel(testModelInfo) {
		t.Fatal("expected test model")
	}
	if modelcatalog.IsTestModel(prodModelInfo) {
		t.Fatal("expected prod model not to be test")
	}
}

func TestFilterVisible(t *testing.T) {
	t.Parallel()
	catalog := []types.ModelInfo{
		{ID: devModel1, Type: modelcatalog.TestCallType, Source: modelcatalog.SourceTest, Enabled: true},
		{ID: devModel100, Type: "gpt-4o", Source: "sms", Enabled: true},
	}
	visible := modelcatalog.FilterVisible(catalog)
	if len(visible) != 1 {
		t.Fatalf("expected 1 visible model, got %d", len(visible))
	}
	if visible[0].Type != "gpt-4o" {
		t.Fatalf("expected gpt-4o, got %s", visible[0].Type)
	}
}
