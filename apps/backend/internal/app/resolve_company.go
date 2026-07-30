package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/seed/contract"
)

// DemoCompanyID is the fixed company ID used in SaaS mode for bootstrap/demo data.
// SaaS mode never goes through setup — this constant is returned directly.
var DemoCompanyID = contract.DefaultCompanyID

const setupCompanyIDKey = "setup_company_id"

// ResolveCompanyID determines the company ID at startup.
//   - SaaS mode → fixed DemoCompanyID (no setup needed)
//   - Local + bootstrap mode (demo/prod/minimal) → fixed DemoCompanyID (seed will create it)
//   - Local + none → read from system_settings; uuid.Nil means not yet initialized
func ResolveCompanyID(ctx context.Context, pool *pgxpool.Pool, cfg config.Config) (uuid.UUID, error) {
	if cfg.SupportSaas {
		return DemoCompanyID, nil
	}
	if cfg.BootstrapNeedsSeed() {
		return DemoCompanyID, nil
	}
	return readSetupCompanyID(ctx, pool)
}

// readSetupCompanyID reads setup_company_id from system_settings.
// Returns uuid.Nil if the key does not exist (not yet initialized).
func readSetupCompanyID(ctx context.Context, pool *pgxpool.Pool) (uuid.UUID, error) {
	var value string
	err := pool.QueryRow(ctx, `SELECT value FROM system_settings WHERE key = $1`, setupCompanyIDKey).Scan(&value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, nil
		}
		return uuid.Nil, fmt.Errorf("read setup_company_id: %w", err)
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse setup_company_id %q: %w", value, err)
	}
	return id, nil
}
