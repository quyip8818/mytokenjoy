package postgres

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/config"
)

//go:embed schema.sql
var schemaSQL string

//go:embed river_schema.sql
var riverSchemaSQL string

// applySchema applies embedded DDL on startup. App tables use CREATE IF NOT EXISTS
// for idempotent bootstrap on empty databases; River schema installs once per DB.
func applySchema(ctx context.Context, pool *pgxpool.Pool, cfg config.Config) error {
	if _, err := pool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS ltree`); err != nil {
		return fmt.Errorf("create ltree extension: %w", err)
	}
	if _, err := pool.Exec(ctx, schemaSQL); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	var riverInstalled bool
	if err := pool.QueryRow(ctx, `SELECT to_regclass('river_job') IS NOT NULL`).Scan(&riverInstalled); err != nil {
		return fmt.Errorf("check river schema: %w", err)
	}
	if !riverInstalled {
		if _, err := pool.Exec(ctx, riverSchemaSQL); err != nil {
			return fmt.Errorf("apply river schema: %w", err)
		}
	}
	if err := applyMonthlyPartitions(ctx, pool, cfg); err != nil {
		return fmt.Errorf("apply partitions: %w", err)
	}
	return nil
}
