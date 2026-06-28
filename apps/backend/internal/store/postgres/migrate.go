package postgres

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func migrate(ctx context.Context, pool *pgxpool.Pool) error {
	entries, err := fs.Glob(migrationFS, "migrations/*.up.sql")
	if err != nil {
		return fmt.Errorf("list migrations: %w", err)
	}
	sort.Strings(entries)
	for _, path := range entries {
		body, err := migrationFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", path, err)
		}
		name := strings.TrimSuffix(strings.TrimPrefix(path, "migrations/"), ".up.sql")
		var exists bool
		if err := pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = 'schema_migrations'
			)
		`).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			if _, err := pool.Exec(ctx, `
				CREATE TABLE schema_migrations (
					version TEXT PRIMARY KEY,
					applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
				)
			`); err != nil {
				return fmt.Errorf("create schema_migrations: %w", err)
			}
		}
		var applied bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`, name).Scan(&applied); err != nil {
			return err
		}
		if applied {
			continue
		}
		if _, err := pool.Exec(ctx, string(body)); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		if _, err := pool.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, name); err != nil {
			return fmt.Errorf("record migration %s: %w", name, err)
		}
	}
	return nil
}
