package postgres

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/store/seed"
)

//go:embed schema.sql
var schemaSQL string

const migrationSQL = `
ALTER TABLE companies ADD COLUMN IF NOT EXISTS authz_revision BIGINT NOT NULL DEFAULT 0;
ALTER TABLE company_recharge_orders ADD COLUMN IF NOT EXISTS display_order_id TEXT NOT NULL DEFAULT '';
ALTER TABLE company_recharge_orders ADD COLUMN IF NOT EXISTS payment_method TEXT NOT NULL DEFAULT '';
ALTER TABLE company_recharge_orders ADD COLUMN IF NOT EXISTS invoice_status TEXT NOT NULL DEFAULT 'none';
UPDATE company_recharge_orders SET display_order_id = id WHERE display_order_id = '';
ALTER TABLE org_integration ADD COLUMN IF NOT EXISTS field_mappings JSONB NOT NULL DEFAULT '[]';
`

func applySchema(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, schemaSQL); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	if _, err := pool.Exec(ctx, migrationSQL); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}

func ensureDemoMemberPasswords(ctx context.Context, pool *pgxpool.Pool, cfg config.Config) error {
	if cfg.IsProdProfile() {
		return nil
	}
	hash := seed.DemoPasswordHash()
	_, err := pool.Exec(ctx, `
		UPDATE members
		SET password_hash = $1, updated_at = NOW()
		WHERE password_hash IS NULL OR password_hash = ''
	`, hash)
	if err != nil {
		return fmt.Errorf("backfill demo member passwords: %w", err)
	}
	return nil
}
