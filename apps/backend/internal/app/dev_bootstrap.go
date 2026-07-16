package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tokenjoy/backend/internal/adapter"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/infra/jobs"
)

// RunDevBootstrap seeds an empty database and synchronously syncs demo platform keys to NewAPI.
func RunDevBootstrap(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	if !cfg.AllowsDevHTTPRoutes() {
		return fmt.Errorf("dev bootstrap requires local deploy env")
	}
	if !cfg.NewAPIEnabled {
		return fmt.Errorf("NEW_API_ENABLED is required for dev bootstrap")
	}

	st, err := openStore(ctx, cfg)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer func() {
		if closer, ok := st.(interface{ Close() }); ok {
			closer.Close()
		}
	}()

	var o options
	o.skipWorker = true
	holder := jobs.NewHolder(jobs.NoopEnqueuer{})
	orgAdmin := adapter.NewOrgRiverAdminHolder(nil)
	registry, err := assembleRegistry(cfg, logger, st, o, holder, orgAdmin)
	if err != nil {
		return fmt.Errorf("assemble registry: %w", err)
	}
	if err := registry.Credentials.BootstrapPlatformIfNeeded(ctx); err != nil {
		return fmt.Errorf("bootstrap platform credentials: %w", err)
	}

	sync, ok := registry.Infra.newAPISync.(*newapisync.NewAPISync)
	if !ok {
		return fmt.Errorf("newapi sync is not configured")
	}
	bootstrapCtx := company.DefaultContext(cfg.LocalCompanyID)
	if err := sync.Bootstrap(bootstrapCtx, cfg.LocalCompanyID); err != nil {
		return fmt.Errorf("bootstrap platform keys: %w", err)
	}

	// Sync wallet_remain → NewAPI user quota so the user has non-zero quota after seed recharge.
	if err := registry.BillingSvc.SyncCompanyWallet(bootstrapCtx, cfg.LocalCompanyID); err != nil {
		logger.Warn("dev bootstrap: wallet sync failed (non-fatal)", "error", err)
	}
	return nil
}
