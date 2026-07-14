package billing

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/newapiunits"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *service) SyncCompanyWallet(ctx context.Context, companyID int64) error {
	co, err := s.store.Company().GetByID(ctx, companyID)
	if err != nil {
		return err
	}
	if co == nil {
		return domain.NotFound("company not found")
	}
	walletUserID, ok := store.ConfiguredNewAPIWalletUserID(co)
	if !ok {
		return ErrWalletNotConfigured
	}
	if s.client == nil || s.wallet == nil {
		return fmt.Errorf("newapi admin client required")
	}
	models, err := s.store.Models().Models(ctx)
	if err != nil {
		return err
	}
	target := newapiunits.ToNewAPIUnits(co.WalletRemain, models, nil)
	current, err := s.wallet.FreshNewAPIUnits(ctx, walletUserID)
	if err != nil {
		return err
	}
	delta := newapiunits.QuotaDelta(target, current)
	if delta == 0 {
		return nil
	}
	if err := s.client.TopUp(ctx, adminport.TopUpInput{
		UserID: walletUserID,
		Quota:  delta,
	}); err != nil {
		return err
	}
	s.wallet.InvalidateNewAPIUnits(walletUserID)
	return nil
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
		walletUserID, ok := store.ConfiguredNewAPIWalletUserID(&co)
		if !ok {
			continue
		}
		quota, err := s.wallet.FreshNewAPIUnits(ctx, walletUserID)
		if err != nil {
			continue
		}
		naPoint := newapiunits.FromNewAPIUnits(quota, models, nil)
		drift := co.WalletRemain - naPoint
		if drift < 0 {
			drift = -drift
		}
		if drift <= common.WalletSyncDriftEpsilon {
			continue
		}
		if err := s.enqueuer.InsertWalletSync(company.WithContext(ctx, company.Context{CompanyID: co.ID}), co.ID); err != nil {
			slog.Warn("wallet drift reconcile: enqueue wallet sync failed", "company_id", co.ID, "err", err)
		}
	}
	return nil
}
