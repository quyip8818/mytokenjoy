package postgres

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schemaSQL string

func applySchema(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS ltree`); err != nil {
		return fmt.Errorf("create ltree extension: %w", err)
	}
	if _, err := pool.Exec(ctx, schemaSQL); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	if err := applyMonthlyPartitions(ctx, pool); err != nil {
		return fmt.Errorf("apply partitions: %w", err)
	}
	return nil
}
