package runtime

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed"
	"github.com/tokenjoy/backend/seed/contract"
)

func ApplyUsageLedger(ctx context.Context, st store.Store, cfg config.Config) error {
	ctx = company.WithContext(ctx, company.Context{CompanyID: contract.DefaultCompanyID})
	entries, _, err := st.Ledger().ListCallSettledPage(ctx, store.LedgerCallFilter{Page: 1, PageSize: 1})
	if err != nil {
		return fmt.Errorf("count usage ledger: %w", err)
	}
	if len(entries) > 0 {
		return nil
	}
	var snap store.Snapshot
	if cfg.BootstrapIsMinimal() {
		snap = seed.LoadMinimal(cfg)
	} else {
		snap = seed.Load(cfg)
	}
	for _, entry := range snap.UsageLedger {
		if _, err := st.Ledger().InsertOnConflict(ctx, entry); err != nil {
			return fmt.Errorf("seed usage ledger %s: %w", entry.ID, err)
		}
	}
	return nil
}
