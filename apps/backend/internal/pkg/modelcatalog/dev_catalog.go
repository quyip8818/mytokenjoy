package modelcatalog

import (
	"strings"

	"github.com/tokenjoy/backend/internal/domain/types"
)

// TestCallType is the local ingest mock call type.
const TestCallType = "test-model"

// IsTestModel returns true if the model type starts with "test-".
func IsTestModel(m types.ModelInfo) bool {
	return strings.HasPrefix(m.Type, "test-")
}

// IsTestOnlyCallType is Gateway-blocked outside DEPLOY_ENV=local and
// allowlist-exempt when local routes are enabled.
func IsTestOnlyCallType(callType string) bool {
	return strings.HasPrefix(callType, "test-")
}
