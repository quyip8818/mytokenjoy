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
