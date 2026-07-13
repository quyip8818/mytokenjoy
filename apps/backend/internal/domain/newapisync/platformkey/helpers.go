package platformkey

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/newapisync/syncdeps"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/newapiunits"
)

func newAPIWalletUserID(ctx context.Context, d syncdeps.Deps) int64 {
	if companyCtx, ok := company.FromContext(ctx); ok && companyCtx.NewAPIWalletUserID > 0 {
		return companyCtx.NewAPIWalletUserID
	}
	companyID := company.CompanyID(ctx)
	co, err := d.Store.Company().GetByID(ctx, companyID)
	if err != nil || co == nil || co.NewAPIWalletUserID == nil {
		return 0
	}
	return *co.NewAPIWalletUserID
}

func capRemainUnits(ctx context.Context, d syncdeps.Deps, remainPoint float64, models []types.ModelInfo, effectiveIDs []int64) int64 {
	allocated := newapiunits.ToNewAPIUnits(remainPoint, models, effectiveIDs)
	if d.Wallet == nil {
		return allocated
	}
	walletID := newAPIWalletUserID(ctx, d)
	if walletID <= 0 {
		return allocated
	}
	walletUnits, err := d.Wallet.AvailableNewAPIUnits(ctx, walletID)
	if err != nil {
		return allocated
	}
	if allocated < walletUnits {
		return allocated
	}
	return walletUnits
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
		if keys[i].ID != platformKeyID {
			continue
		}
		keys[i].FullKey = &fullKey
		keys[i].KeyPrefix = newAPIPlatformKeyPrefix(fullKey)
		return d.Store.Keys().SetPlatformKeys(ctx, keys)
	}
	return fmt.Errorf("platform key not found: %s", platformKeyID)
}
