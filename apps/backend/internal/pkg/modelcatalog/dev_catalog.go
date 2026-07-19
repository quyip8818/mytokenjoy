package modelcatalog

import (
	"strings"

	"github.com/google/uuid"
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

// ModelLimitsCallTypes builds NewAPI token model_limits from a Platform Key whitelist.
// Pass includeTestCatalog=true only when DEPLOY_ENV=local (AllowsDevHTTPRoutes).
func ModelLimitsCallTypes(catalog []types.ModelInfo, allowedIDs []uuid.UUID, includeTestCatalog bool) []string {
	if !includeTestCatalog {
		return CallTypesForIDs(catalog, allowedIDs)
	}
	return CallTypesForIDs(catalog, withTestModelIDs(catalog, allowedIDs))
}

func withTestModelIDs(catalog []types.ModelInfo, ids []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(ids)+4)
	out := make([]uuid.UUID, 0, len(ids)+4)
	for _, mid := range ids {
		if _, ok := seen[mid]; ok {
			continue
		}
		seen[mid] = struct{}{}
		out = append(out, mid)
	}
	for _, item := range catalog {
		if !item.Enabled || !IsTestModel(item) {
			continue
		}
		if _, ok := seen[item.ID]; ok {
			continue
		}
		seen[item.ID] = struct{}{}
		out = append(out, item.ID)
	}
	return out
}
