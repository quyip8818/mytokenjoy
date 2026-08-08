package modelcatalog

import "github.com/tokenjoy/backend/internal/domain/types"

// TestCallType is the local ingest mock call type.
const TestCallType = "test-model"

// SourceTest marks models injected by seed for full-path testing.
const SourceTest = "test"

// IsTestCallType returns true if the model type string is the dev/test mock.
func IsTestCallType(modelType string) bool {
	return modelType == TestCallType
}

// IsTestModel returns true if the model's source is "test".
func IsTestModel(m types.ModelInfo) bool {
	return m.Source == SourceTest
}

// FilterVisible returns only non-test models (source != "test").
func FilterVisible(models []types.ModelInfo) []types.ModelInfo {
	out := make([]types.ModelInfo, 0, len(models))
	for i := range models {
		if !IsTestModel(models[i]) {
			out = append(out, models[i])
		}
	}
	return out
}

// FilterNotDeprecated returns only non-deprecated models (deprecated == false).
func FilterNotDeprecated(models []types.ModelInfo) []types.ModelInfo {
	out := make([]types.ModelInfo, 0, len(models))
	for i := range models {
		if !models[i].Deprecated {
			out = append(out, models[i])
		}
	}
	return out
}
