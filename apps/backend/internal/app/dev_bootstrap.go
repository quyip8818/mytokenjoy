package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/adapter/enqueue"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/integration/newapisync"
)

// RunDevBootstrap seeds an empty database and synchronously syncs demo platform keys to NewAPI.
func RunDevBootstrap(ctx context.Context, cfg config.Config, logger *slog.Logger) error {
	if !cfg.AllowsDevHTTPRoutes() {
		return fmt.Errorf("dev bootstrap requires local deploy env")
	}
	if !cfg.NewAPIEnabled {
		return fmt.Errorf("NEW_API_ENABLED is required for dev bootstrap")
	}

	// Resolve CompanyID before opening store (seed.Init needs it).
	// Dev bootstrap always uses DemoCompanyID — it seeds a fresh DB without setup flow.
	if cfg.CompanyID == (uuid.UUID{}) {
		cfg.CompanyID = DemoCompanyID
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
	orgAdmin := enqueue.NewOrgRiverAdminHolder(nil)
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
	bootstrapCtx := company.DefaultContext(cfg.CompanyID)
	if err := sync.Bootstrap(bootstrapCtx, cfg.CompanyID); err != nil {
		logger.Warn("platform key sync failed (non-fatal, keys will sync on next request)", "error", err)
	}

	return nil
}
