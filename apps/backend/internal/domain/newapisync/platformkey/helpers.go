package platformkey

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/newapisync/syncdeps"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
	"github.com/tokenjoy/backend/internal/pkg/newapiunits"
	"github.com/tokenjoy/backend/internal/store"
)

func newAPIWalletCompanyID(ctx context.Context, d syncdeps.Deps) (int64, error) {
	if companyCtx, ok := company.FromContext(ctx); ok && companyCtx.NewAPIWalletCompanyID > 0 {
		return companyCtx.NewAPIWalletCompanyID, nil
	}
	companyID := company.CompanyID(ctx)
	co, err := d.Store.Company().GetByID(ctx, companyID)
	if err != nil {
		return 0, err
	}
	if co == nil {
		return 0, nil
	}
	id, ok := store.ConfiguredNewAPIWalletCompanyID(co)
	if !ok {
		return 0, nil
	}
	return id, nil
}

func newAPIPlatformKeyPrefix(fullKey string) string {
	prefix := fullKey
	if len(prefix) > 12 {
		prefix = prefix[:12] + "..."
	}
	return prefix
}

func persistPlatformKeySecret(ctx context.Context, d syncdeps.Deps, platformKeyID uuid.UUID, fullKey string) error {
	keys, err := d.Store.Keys().PlatformKeys(ctx)
	if err != nil {
		return err
	}
	for i := range keys {
		if keys[i].ID == platformKeyID {
			keys[i].FullKey = &fullKey
			keys[i].KeyPrefix = newAPIPlatformKeyPrefix(fullKey)
			return d.Store.Keys().SetPlatformKeys(ctx, keys)
		}
	}
	return fmt.Errorf("platform key not found: %s", platformKeyID)
}

func resolveModelLimits(d syncdeps.Deps, models []types.ModelInfo, keyWhitelist, deptAllowed []uuid.UUID) (effectiveIDs []uuid.UUID, callTypes []string) {
	effectiveIDs = newapiunits.EffectiveWhitelistIDs(keyWhitelist, deptAllowed)
	callTypes = modelcatalog.ModelLimitsCallTypes(models, effectiveIDs, d.Cfg.AllowsDevHTTPRoutes())
	return effectiveIDs, callTypes
}
