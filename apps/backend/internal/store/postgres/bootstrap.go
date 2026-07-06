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
	return nil
}
