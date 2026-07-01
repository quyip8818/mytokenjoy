package relay

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
)

func (l *TokenLifecycle) walletUserID(ctx context.Context) int64 {
	if companyCtx, ok := company.FromContext(ctx); ok && companyCtx.NewAPIWalletAccountID > 0 {
		return companyCtx.NewAPIWalletAccountID
	}
	companyID := company.CompanyID(ctx)
	company, err := l.store.Company().GetByID(ctx, companyID)
	if err != nil || company == nil || company.NewAPIWalletAccountID == nil {
		return 0
	}
	return *company.NewAPIWalletAccountID
}

func (l *TokenLifecycle) capRemainUnits(ctx context.Context, remainCNY float64, models []types.ModelInfo, effective []string) int64 {
	allocated := newapi.ToNewAPIUnits(remainCNY, models, effective)
	if l.wallet == nil {
		return allocated
	}
	walletID := l.walletUserID(ctx)
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
