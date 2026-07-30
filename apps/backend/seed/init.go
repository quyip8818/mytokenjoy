package seed

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/bootstrap"
	"github.com/tokenjoy/backend/seed/runtime"
)

// Init is the single entry point for all database data initialization.
// Called from store/postgres.New after schema DDL is applied.
//
// Sequence:
//  1. bootstrap.ApplyBootstrap — currencies, companies, permissions, roles, org (idempotent)
//  2. SaaS mode + empty DB → apply demo snapshot data
//
// Local mode: bootstrap only (company created by setup flow, no demo data).
func Init(ctx context.Context, pool *pgxpool.Pool, st store.Store, cfg config.Config) error {
	// Load bootstrap config from file or defaults.
	bsCfg, err := bootstrap.LoadConfig(os.Getenv("BOOTSTRAP_CONFIG_PATH"))
	if err != nil {
		return fmt.Errorf("seed init: %w", err)
	}

	// 1. Always apply bootstrap (idempotent): currencies, permissions, roles, org, models.
	if err := bootstrap.ApplyBootstrap(ctx, pool, cfg, bsCfg); err != nil {
		return fmt.Errorf("seed init: %w", err)
	}

	// 2. SaaS mode: seed demo data on first boot (empty DB).
	if cfg.SupportSaas {
		empty, err := isDatabaseEmpty(ctx, pool)
		if err != nil {
			return err
		}
		if empty {
			if err := applySeedData(ctx, pool, st, cfg); err != nil {
				return err
			}
		}
	}

	return nil
}

func applySeedData(ctx context.Context, pool *pgxpool.Pool, st store.Store, cfg config.Config) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin seed tx: %w", err)
	}
	defer tx.Rollback(ctx)

	snap := Load(cfg)
	if err := ApplyTables(ctx, tx, snap); err != nil {
		return err
	}
	// Seed includes platform models — set initial catalog version so sync clients see them.
	if _, err := tx.Exec(ctx, `INSERT INTO system_settings (key, value) VALUES ('catalog.models_version', '1') ON CONFLICT (key) DO NOTHING`); err != nil {
		return fmt.Errorf("seed catalog.models_version: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit seed tx: %w", err)
	}
	// Runtime demo data (usage ledger projections etc.)
	return runtime.ApplyDemo(ctx, st, cfg)
}

func isDatabaseEmpty(ctx context.Context, pool *pgxpool.Pool) (bool, error) {
	var count int
	err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM members`).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("count members: %w", err)
	}
	return count == 0, nil
}
