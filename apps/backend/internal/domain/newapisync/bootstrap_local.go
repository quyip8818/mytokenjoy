package newapisync

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tokenjoy/backend/internal/domain/company"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

// BootstrapUnsyncedPlatformKeys enqueues NewAPI create jobs for demo seed platform keys
// inserted directly by seed apply. Local-only; same outbox as production.
func (l *NewAPISync) BootstrapUnsyncedPlatformKeys(ctx context.Context, companyID int64) error {
	if !l.Enabled() || !l.cfg.AllowsDevHTTPRoutes() {
		return nil
	}
	ctx = company.WithDefaultCompany(ctx, companyID)

	platformKeys, err := l.store.Keys().PlatformKeys(ctx)
	if err != nil {
		return err
	}
	budgetCtx, err := pkgbudget.LoadBudgetContext(
		ctx, l.store.BudgetConsumed(), l.store.Org(), l.store.Budget(), l.store.Keys(), l.cfg.Clock(),
	)
	if err != nil {
		return fmt.Errorf("load budget context: %w", err)
	}

	for _, key := range platformKeys {
		if key.Status != "active" {
			continue
		}
		mapping, err := l.mappings.GetMappingByPlatformKeyID(ctx, key.ID)
		if err != nil {
			return err
		}
		if mapping != nil && mapping.SyncStatus == store.MappingSyncStatusSynced && mapping.NewAPIKeyID != nil {
			continue
		}
		departmentID := departmentIDForPlatformKey(key, budgetCtx)
		if departmentID == "" {
			continue
		}
		if err := l.SyncCreatePlatformKey(ctx, key, departmentID); err != nil {
			return fmt.Errorf("bootstrap platform key %s: %w", key.ID, err)
		}
	}
	if err := l.repairStalePlatformKeyHashes(ctx); err != nil {
		slog.Default().Warn("repair stale platform key hashes failed", "error", err)
	}
	return nil
}
