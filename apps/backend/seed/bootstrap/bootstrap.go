package bootstrap

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tokenjoy/backend/internal/config"
)

// TableWriter is the minimal interface for executing SQL statements.
// Satisfied by *pgxpool.Pool, pgx.Tx, etc.
type TableWriter interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

// ApplyBootstrap writes the minimum data required for any environment to start.
// All operations are idempotent (ON CONFLICT DO NOTHING).
func ApplyBootstrap(ctx context.Context, exec TableWriter, appCfg config.Config, bsCfg Config) error {
	companyID := appCfg.LocalCompanyID
	tokenJoyID := appCfg.TokenJoyCompanyID

	if err := insertCurrencies(ctx, exec, bsCfg); err != nil {
		return fmt.Errorf("bootstrap currencies: %w", err)
	}
	if err := insertCompanies(ctx, exec, appCfg, bsCfg, tokenJoyID, companyID); err != nil {
		return fmt.Errorf("bootstrap companies: %w", err)
	}
	if err := insertPermissions(ctx, exec); err != nil {
		return fmt.Errorf("bootstrap permissions: %w", err)
	}
	if err := seedGlobalPresetRoles(ctx, exec); err != nil {
		return fmt.Errorf("bootstrap roles: %w", err)
	}
	if err := insertRootOrg(ctx, exec, companyID, appCfg, bsCfg); err != nil {
		return fmt.Errorf("bootstrap org: %w", err)
	}
	if err := insertModels(ctx, exec, companyID, bsCfg); err != nil {
		return fmt.Errorf("bootstrap models: %w", err)
	}
	if err := insertTenantBackgroundState(ctx, exec, tokenJoyID, companyID); err != nil {
		return fmt.Errorf("bootstrap tenant_background_state: %w", err)
	}
	return nil
}

func insertTenantBackgroundState(ctx context.Context, exec TableWriter, tokenJoyID, companyID uuid.UUID) error {
	for _, id := range []uuid.UUID{tokenJoyID, companyID} {
		if _, err := exec.Exec(ctx, `
			INSERT INTO tenant_background_state (company_id) VALUES ($1)
			ON CONFLICT (company_id) DO NOTHING
		`, id); err != nil {
			return fmt.Errorf("insert tenant_background_state %s: %w", id, err)
		}
	}
	return nil
}
