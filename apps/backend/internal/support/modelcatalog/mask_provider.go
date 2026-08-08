package modelcatalog

import "github.com/tokenjoy/backend/internal/domain/types"

// DisplayProvider is the unified provider name shown to tenant users.
// Platform models are managed by multiple upstream providers (openai, volcengine, etc.)
// but customers only see "tokenjoy" as the single provider.
const DisplayProvider = "tokenjoy"

// MaskProviderForTenant replaces the real upstream provider with DisplayProvider
// for platform-managed models before returning to tenant-facing APIs.
//
// DB stores the real provider (Platform Admin needs it for internal management).
// This function masks it at the API response layer — customers never see upstream details.
//
// Only call in tenant-facing endpoints (GET /api/models).
// Never call for: Platform Admin API, CatalogSync internals, Gateway precheck.
func MaskProviderForTenant(models []types.ModelInfo) {
	for i := range models {
		if models[i].Source == "platform" || models[i].Source == "seed" {
			models[i].Provider = DisplayProvider
		}
	}
}

// ponytail: single function, zero-alloc (mutates slice in place).
// Upgrade path: if we need per-company branding, add a brandName parameter.
