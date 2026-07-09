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

func applySchema(ctx context.Context, pool *pgxpool.Pool, cfg config.Config) error {
	if _, err := pool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS ltree`); err != nil {
		return fmt.Errorf("create ltree extension: %w", err)
	}
	if _, err := pool.Exec(ctx, schemaSQL); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	if err := applyMonthlyPartitions(ctx, pool, cfg); err != nil {
		return fmt.Errorf("apply partitions: %w", err)
	}
	return nil
}
