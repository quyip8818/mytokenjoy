package seed

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
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
//  2. bootstrap.ReconcilePresetRoles — ensure grants match manifest (every startup)
//  3. if demo/minimal + empty DB → apply demo/minimal snapshot data
func Init(ctx context.Context, pool *pgxpool.Pool, st store.Store, cfg config.Config) error {
	if cfg.BootstrapIsNone() {
		// none mode: no initialization. Error out if DB is empty.
		empty, err := isDatabaseEmpty(ctx, pool)
		if err != nil {
			return err
		}
		if empty {
			return fmt.Errorf("database empty: set BOOTSTRAP_MODE=prod|minimal|demo or run migrations externally")
		}
		return nil
	}

	// Load bootstrap config from file or defaults.
	bsCfg, err := bootstrap.LoadConfig(os.Getenv("BOOTSTRAP_CONFIG_PATH"))
	if err != nil {
		return fmt.Errorf("seed init: %w", err)
	}

	// 1. Always apply bootstrap (idempotent).
	if err := bootstrap.ApplyBootstrap(ctx, pool, cfg, bsCfg); err != nil {
		return fmt.Errorf("seed init: %w", err)
	}

	// 2. Reconcile preset roles for all companies (every startup).
	companyIDs, err := listCompanyIDs(ctx, pool)
	if err != nil {
		return fmt.Errorf("seed init: list companies: %w", err)
	}
	if err := bootstrap.ReconcilePresetRoles(ctx, pool, companyIDs); err != nil {
		return fmt.Errorf("seed init: %w", err)
	}

	// 3. Conditionally apply seed data on empty DB.
	if cfg.BootstrapIsMinimal() || cfg.BootstrapIsDemo() {
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

	var snap store.Snapshot
	switch {
	case cfg.SeedIsEmpty():
		snap = LoadEmpty(cfg)
	case cfg.SeedIsMinimal() || cfg.BootstrapIsMinimal():
		snap = LoadMinimal(cfg)
	default:
		snap = Load(cfg)
	}

	if err := ApplyTables(ctx, tx, snap); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit seed tx: %w", err)
	}
	// Runtime demo data (usage ledger projections etc.) only for full seed.
	if cfg.BootstrapIsDemo() && cfg.SeedIsFull() {
		return runtime.ApplyDemo(ctx, st, cfg)
	}
	return nil
}

func listCompanyIDs(ctx context.Context, pool *pgxpool.Pool) ([]uuid.UUID, error) {
	rows, err := pool.Query(ctx, `SELECT id FROM companies`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func isDatabaseEmpty(ctx context.Context, pool *pgxpool.Pool) (bool, error) {
	var count int
	err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM members`).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("count members: %w", err)
	}
	return count == 0, nil
}
