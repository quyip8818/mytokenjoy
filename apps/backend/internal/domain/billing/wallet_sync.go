package billing

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

func (s *service) SyncCompanyWallet(ctx context.Context, companyID int64) error {
	co, err := s.store.Company().GetByID(ctx, companyID)
	if err != nil || co == nil || co.NewAPIWalletUserID == nil {
		return fmt.Errorf("company wallet not configured")
	}
	if !s.cfg.NewAPIEnabled || s.client == nil {
		return fmt.Errorf("newapi sync required but newapi is disabled")
	}
	models, err := s.store.Models().Models(ctx)
	if err != nil {
		return err
	}
	target := newapi.ToQuotaUnits(co.BalancePoint, models, nil)
	current, err := s.wallet.AvailableQuota(ctx, *co.NewAPIWalletUserID)
	if err != nil {
		return err
	}
	delta := target - current
	if delta == 0 {
		return nil
	}
	if delta > 0 {
		return s.client.TopUp(ctx, newapi.TopUpRequest{
			UserID: *co.NewAPIWalletUserID,
			Quota:  delta,
			Remark: "wallet_sync",
		})
	}
	return s.client.TopUp(ctx, newapi.TopUpRequest{
		UserID: *co.NewAPIWalletUserID,
		Quota:  delta,
		Remark: "wallet_sync_decrease",
	})
}

func (s *service) ReconcileWalletDrift(ctx context.Context) error {
	if !s.cfg.NewAPIEnabled || s.client == nil || s.wallet == nil {
		return nil
	}
	companies, err := s.store.Company().List(ctx)
	if err != nil {
		return err
	}
	models, err := s.store.Models().Models(ctx)
	if err != nil {
		return err
	}
	for _, co := range companies {
		if co.NewAPIWalletUserID == nil {
			continue
		}
		quota, err := s.wallet.AvailableQuota(ctx, *co.NewAPIWalletUserID)
		if err != nil {
			continue
		}
		naPoint := newapi.FromNewAPIUnits(quota, models, nil)
		drift := co.BalancePoint - naPoint
		if drift < 0 {
			drift = -drift
		}
		if drift <= common.WalletSyncDriftEpsilon {
			continue
		}
		if s.enqueueSync != nil {
			_ = s.enqueueSync(company.WithContext(ctx, company.Context{CompanyID: co.ID}), co.ID)
		}
	}
	return nil
}
