package modelcatalog

import (
	"strings"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
)

// DevCallTypeLocalTest is the local ingest mock call type.
const DevCallTypeLocalTest = "dev-local-test"

// IsDevModel returns true if the model type starts with "dev-".
func IsDevModel(m types.ModelInfo) bool {
	return strings.HasPrefix(m.Type, "dev-")
}

// IsLocalOnlyCallType is Gateway-blocked outside DEPLOY_ENV=local and
// allowlist-exempt when local routes are enabled.
func IsLocalOnlyCallType(callType string) bool {
	return strings.HasPrefix(callType, "dev-")
}

// ModelLimitsCallTypes builds NewAPI token model_limits from a Platform Key whitelist.
// Pass includeDevCatalog=true only when DEPLOY_ENV=local (AllowsDevHTTPRoutes).
func ModelLimitsCallTypes(catalog []types.ModelInfo, allowedIDs []uuid.UUID, includeDevCatalog bool) []string {
	if !includeDevCatalog {
		return CallTypesForIDs(catalog, allowedIDs)
	}
	return CallTypesForIDs(catalog, withDevModelIDs(catalog, allowedIDs))
}

func withDevModelIDs(catalog []types.ModelInfo, ids []uuid.UUID) []uuid.UUID {
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
		if !item.Enabled || !IsDevModel(item) {
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
