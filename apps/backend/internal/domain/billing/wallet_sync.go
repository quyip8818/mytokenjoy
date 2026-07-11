package billing

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/newapiunits"
)

func (s *service) SyncCompanyWallet(ctx context.Context, companyID int64) error {
	co, err := s.store.Company().GetByID(ctx, companyID)
	if err != nil {
		return err
	}
	if co == nil || co.NewAPIWalletUserID == nil {
		return ErrWalletNotConfigured
	}
	if s.client == nil {
		return fmt.Errorf("newapi admin client required")
	}
	models, err := s.store.Models().Models(ctx)
	if err != nil {
		return err
	}
	target := newapiunits.ToQuotaUnits(co.BalancePoint, models, nil)
	current, err := s.wallet.AvailableQuota(ctx, *co.NewAPIWalletUserID)
	if err != nil {
		return err
	}
	delta := target - current
	if delta == 0 {
		return nil
	}
	if delta > 0 {
		return s.client.TopUp(ctx, adminport.TopUpInput{
			UserID: *co.NewAPIWalletUserID,
			Quota:  delta,
			Remark: "wallet_sync",
		})
	}
	return s.client.TopUp(ctx, adminport.TopUpInput{
		UserID: *co.NewAPIWalletUserID,
		Quota:  delta,
		Remark: "wallet_sync_decrease",
	})
}

func (s *service) ReconcileWalletDrift(ctx context.Context) error {
	if s.client == nil || s.wallet == nil {
		return fmt.Errorf("newapi wallet client required")
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
		naPoint := newapiunits.FromNewAPIUnits(quota, models, nil)
		drift := co.BalancePoint - naPoint
		if drift < 0 {
			drift = -drift
		}
		if drift <= common.WalletSyncDriftEpsilon {
			continue
		}
		if s.enqueueSync != nil {
			if err := s.enqueueSync(company.WithContext(ctx, company.Context{CompanyID: co.ID}), co.ID); err != nil {
				slog.Warn("wallet drift reconcile: enqueue wallet sync failed", "company_id", co.ID, "err", err)
			}
		}
	}
	return nil
}
