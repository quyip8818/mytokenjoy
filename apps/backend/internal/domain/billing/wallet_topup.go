package billing

import (
	"context"
	"log/slog"

	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/store"
)

// topUpNewAPIQuota tops up the NewAPI user quota by delta after a successful credit.
// Best-effort: failures are logged but do not fail the recharge transaction.
func (s *service) topUpNewAPIQuota(ctx context.Context, delta int64) {
	if s.adminClient == nil || delta <= 0 {
		return
	}
	companyID := company.CompanyID(ctx)
	co, err := s.store.Company().GetByID(ctx, companyID)
	if err != nil {
		slog.Default().Warn("topUpNewAPIQuota: get company failed", "company_id", companyID, "error", err)
		return
	}
	walletUserID, ok := store.ConfiguredNewAPIWalletUserID(co)
	if !ok {
		return
	}
	if err := s.adminClient.TopUp(ctx, adminport.TopUpInput{
		UserID: walletUserID,
		Quota:  delta,
	}); err != nil {
		slog.Default().Warn("topUpNewAPIQuota: TopUp failed", "company_id", companyID, "wallet_user_id", walletUserID, "delta", delta, "error", err)
	}
}
