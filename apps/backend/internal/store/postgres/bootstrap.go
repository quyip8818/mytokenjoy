package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

func ensureBootstrapCompany(ctx context.Context, pool *pgxpool.Pool, cfg config.Config) error {
	if err := ensureBootstrapCurrencies(ctx, pool); err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO companies (id, name, type, status, newapi_wallet_username)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO NOTHING
	`, cfg.TokenJoyCompanyID, "TokenJoy", store.CompanyTypeTesting, store.CompanyStatusActive,
		company.WalletUsername(cfg.TokenJoyCompanyID)); err != nil {
		return fmt.Errorf("bootstrap tokenjoy company: %w", err)
	}

	companyID := cfg.LocalCompanyID
	name := cfg.ResolvedCompanyName()
	companyType := store.CompanyTypeSelfhosted
	if cfg.SupportSaas {
		companyType = store.CompanyTypeTesting
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO companies (id, name, type, status, newapi_wallet_username)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO NOTHING
	`, companyID, name, companyType, store.CompanyStatusActive,
		company.WalletUsername(companyID)); err != nil {
		return fmt.Errorf("bootstrap company: %w", err)
	}
	return nil
}

func ensureBootstrapCurrencies(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO currencies (currency, quota_per_unit, enabled)
		VALUES ($1, $2, TRUE)
		ON CONFLICT (currency) DO NOTHING
	`, common.DefaultBillingCurrency, common.DefaultQuotaPerUnit)
	if err != nil {
		return fmt.Errorf("bootstrap currencies: %w", err)
	}
	return nil
}
