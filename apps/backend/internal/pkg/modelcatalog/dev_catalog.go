package modelcatalog

import "github.com/tokenjoy/backend/internal/domain/types"

// ProdCatalogModelIDStart is the first production catalog model_id.
// IDs below it are local/dev-only (seed mock models).
const ProdCatalogModelIDStart int64 = 100

// DevCallTypeLocalTest is the local ingest mock call type (dev-mock-llm).
const DevCallTypeLocalTest = "local-test-model"

func IsDevCatalogModelID(id int64) bool {
	return id > 0 && id < ProdCatalogModelIDStart
}

// IsLocalOnlyCallType is Gateway-blocked outside DEPLOY_ENV=local and
// allowlist-exempt when local routes are enabled.
func IsLocalOnlyCallType(callType string) bool {
	return callType == DevCallTypeLocalTest
}

// ModelLimitsCallTypes builds NewAPI token model_limits from a Platform Key whitelist.
// Pass includeDevCatalog=true only when DEPLOY_ENV=local (AllowsDevHTTPRoutes).
func ModelLimitsCallTypes(catalog []types.ModelInfo, allowedIDs []int64, includeDevCatalog bool) []string {
	if !includeDevCatalog {
		return CallTypesForIDs(catalog, allowedIDs)
	}
	return CallTypesForIDs(catalog, withDevCatalogIDs(catalog, allowedIDs))
}

func withDevCatalogIDs(catalog []types.ModelInfo, ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids)+4)
	out := make([]int64, 0, len(ids)+4)
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	for _, item := range catalog {
		if !item.Enabled || !IsDevCatalogModelID(item.ModelID) {
			continue
		}
		if _, ok := seen[item.ModelID]; ok {
			continue
		}
		seen[item.ModelID] = struct{}{}
		out = append(out, item.ModelID)
	}
	return out
}
