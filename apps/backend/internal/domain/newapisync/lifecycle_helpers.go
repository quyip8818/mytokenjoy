package newapisync

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
)

func (l *NewAPISync) newAPIWalletUserID(ctx context.Context) int64 {
	if companyCtx, ok := company.FromContext(ctx); ok && companyCtx.NewAPIWalletUserID > 0 {
		return companyCtx.NewAPIWalletUserID
	}
	companyID := company.CompanyID(ctx)
	company, err := l.store.Company().GetByID(ctx, companyID)
	if err != nil || company == nil || company.NewAPIWalletUserID == nil {
		return 0
	}
	return *company.NewAPIWalletUserID
}

func (l *NewAPISync) capRemainUnits(ctx context.Context, remainCNY float64, models []types.ModelInfo, effectiveIDs []int64) int64 {
	allocated := newapi.ToNewAPIUnits(remainCNY, models, effectiveIDs)
	if l.wallet == nil {
		return allocated
	}
	walletID := l.newAPIWalletUserID(ctx)
	if walletID <= 0 {
		return allocated
	}
	walletUnits, err := l.wallet.AvailableQuota(ctx, walletID)
	if err != nil {
		return allocated
	}
	if allocated < walletUnits {
		return allocated
	}
	return walletUnits
}

func findPlatformKey(keys []types.PlatformKey, id string) (types.PlatformKey, bool) {
	for _, key := range keys {
		if key.ID == id {
			return key, true
		}
	}
	return types.PlatformKey{}, false
}

func newAPIPlatformKeyPrefix(fullKey string) string {
	prefix := fullKey
	if len(prefix) > 12 {
		prefix = prefix[:12] + "..."
	}
	return prefix
}
