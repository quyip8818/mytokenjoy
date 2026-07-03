package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/store"
)

func ensureBootstrapCompany(ctx context.Context, pool *pgxpool.Pool, cfg config.Config) error {
	companyID := cfg.DefaultCompanyID
	name := cfg.ResolvedCompanyName()
	if _, err := pool.Exec(ctx, `
		INSERT INTO companies (id, slug, name, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING
	`, companyID, config.DefaultCompanySlug, name, store.CompanyStatusActive); err != nil {
		return fmt.Errorf("bootstrap company: %w", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO relay_sync_cursors (company_id, last_log_id)
		VALUES ($1, 0)
		ON CONFLICT DO NOTHING
	`, companyID); err != nil {
		return fmt.Errorf("bootstrap relay sync cursor: %w", err)
	}
	return nil
}
