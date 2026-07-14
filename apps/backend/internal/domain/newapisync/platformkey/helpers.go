package platformkey

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/newapisync/syncdeps"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
	"github.com/tokenjoy/backend/internal/pkg/newapiunits"
	"github.com/tokenjoy/backend/internal/store"
)

func newAPIWalletUserID(ctx context.Context, d syncdeps.Deps) (int64, error) {
	if companyCtx, ok := company.FromContext(ctx); ok && companyCtx.NewAPIWalletUserID > 0 {
		return companyCtx.NewAPIWalletUserID, nil
	}
	companyID := company.CompanyID(ctx)
	co, err := d.Store.Company().GetByID(ctx, companyID)
	if err != nil {
		return 0, err
	}
	if co == nil {
		return 0, nil
	}
	id, ok := store.ConfiguredNewAPIWalletUserID(co)
	if !ok {
		return 0, nil
	}
	return id, nil
}

func capRemainUnits(ctx context.Context, d syncdeps.Deps, remainPoint float64, models []types.ModelInfo, effectiveIDs []int64) (int64, error) {
	allocated := newapiunits.ToNewAPIUnits(remainPoint, models, effectiveIDs)
	if d.Wallet == nil {
		return allocated, nil
	}
	walletID, err := newAPIWalletUserID(ctx, d)
	if err != nil {
		return 0, err
	}
	if walletID <= 0 {
		return allocated, nil
	}
	walletUnits, err := d.Wallet.AvailableNewAPIUnits(ctx, walletID)
	if err != nil {
		return 0, err
	}
	if allocated < walletUnits {
		return allocated, nil
	}
	return walletUnits, nil
}

func newAPIPlatformKeyPrefix(fullKey string) string {
	prefix := fullKey
	if len(prefix) > 12 {
		prefix = prefix[:12] + "..."
	}
	return prefix
}

func persistPlatformKeySecret(ctx context.Context, d syncdeps.Deps, platformKeyID, fullKey string) error {
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

func resolveModelLimits(d syncdeps.Deps, models []types.ModelInfo, keyWhitelist, deptAllowed []int64) (effectiveIDs []int64, callTypes []string) {
	effectiveIDs = newapiunits.EffectiveWhitelistIDs(keyWhitelist, deptAllowed)
	callTypes = modelcatalog.ModelLimitsCallTypes(models, effectiveIDs, d.Cfg.AllowsDevHTTPRoutes())
	return effectiveIDs, callTypes
}
