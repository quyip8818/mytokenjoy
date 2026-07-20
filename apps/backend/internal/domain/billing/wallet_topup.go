package billing

import (
	"context"
	"log/slog"

	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/company"
)

// topUpNewAPIQuota tops up the NewAPI company quota by delta after a successful credit.
// Best-effort: failures are logged but do not fail the recharge transaction.
func (s *service) topUpNewAPIQuota(ctx context.Context, delta int64) {
	if s.adminClient == nil || delta <= 0 {
		return
	}
	walletCompanyID, ok := company.ResolveNewAPIWalletCompanyID(ctx, s.store.Company())
	if !ok {
		return
	}
	if err := s.adminClient.TopUp(ctx, adminport.TopUpInput{
		CompanyID: walletCompanyID,
		Quota:     delta,
	}); err != nil {
		slog.Default().Warn("topUpNewAPIQuota: TopUp failed", "company_id", company.CompanyID(ctx), "wallet_company_id", walletCompanyID, "delta", delta, "error", err)
	}
}
