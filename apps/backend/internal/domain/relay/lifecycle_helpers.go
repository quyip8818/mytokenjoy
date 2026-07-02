package relay

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
)

func (l *TokenLifecycle) newAPIWalletUserID(ctx context.Context) int64 {
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

func (l *TokenLifecycle) capRemainUnits(ctx context.Context, remainCNY float64, models []types.ModelInfo, effective []string) int64 {
	allocated := newapi.ToNewAPIUnits(remainCNY, models, effective)
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
