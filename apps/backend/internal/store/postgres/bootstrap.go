package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/config"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

func ensureBootstrapCompany(ctx context.Context, pool *pgxpool.Pool, cfg config.Config) error {
	if err := ensureBootstrapCurrencies(ctx, pool); err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO companies (id, slug, name, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING
	`, cfg.TokenJoyCompanyID, "tokenjoy", "TokenJoy", store.CompanyStatusActive); err != nil {
		return fmt.Errorf("bootstrap tokenjoy company: %w", err)
	}

	companyID := cfg.LocalCompanyID
	name := cfg.ResolvedCompanyName()
	if _, err := pool.Exec(ctx, `
		INSERT INTO companies (id, slug, name, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING
	`, companyID, config.DefaultCompanySlug, name, store.CompanyStatusActive); err != nil {
		return fmt.Errorf("bootstrap company: %w", err)
	}
	return validateCompanyIDsForMode(ctx, pool, cfg)
}

func ensureBootstrapCurrencies(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO currencies (currency, points_per_unit, enabled)
		VALUES ('CNY', $1, TRUE)
		ON CONFLICT (currency) DO NOTHING
	`, common.DefaultPointsPerUnit)
	if err != nil {
		return fmt.Errorf("bootstrap currencies: %w", err)
	}
	return nil
}

func validateCompanyIDsForMode(ctx context.Context, pool *pgxpool.Pool, cfg config.Config) error {
	if cfg.SupportSaas {
		return nil
	}
	rows, err := pool.Query(ctx, `
		SELECT id FROM companies WHERE id <> $1 AND id >= $2
	`, cfg.TokenJoyCompanyID, domaincompany.SaaSMinCompanyID)
	if err != nil {
		return fmt.Errorf("validate company ids: %w", err)
	}
	defer rows.Close()
	var invalid []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return err
		}
		invalid = append(invalid, id)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(invalid) > 0 {
		return fmt.Errorf("SUPPORT_SAAS=false but found SaaS-range company ids: %v", invalid)
	}
	return nil
}
