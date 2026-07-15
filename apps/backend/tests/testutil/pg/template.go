//go:build testhook

package pg

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/store/postgres"
)

const testTemplateVersion = 30 // bump when schema.sql (incl. River DDL), seed data, or clone policy changes

var (
	templateOnce sync.Once
	templateErr  error
	clonePlan    ClonePlan
)

func EnsureTemplate(ctx context.Context, baseURL string, templateCfg config.Config) error {
	templateOnce.Do(func() {
		pool, err := EnsureAdminPool(ctx, baseURL)
		if err != nil {
			templateErr = fmt.Errorf("connect postgres for template: %w", err)
			return
		}
		templateErr = withTemplateLock(ctx, pool, func() error {
			cleanupOrphanTestSchemas(ctx, pool)
			if version, ok := readTemplateVersion(ctx, pool); ok && version == testTemplateVersion {
				clonePlan, err = LoadClonePlan(ctx, pool, testTemplateSchema)
				return err
			}
			if err := buildTestTemplate(ctx, baseURL, templateCfg); err != nil {
				return err
			}
			if err := markTemplateVersion(ctx, pool); err != nil {
				return err
			}
			clonePlan, err = LoadClonePlan(ctx, pool, testTemplateSchema)
			return err
		})
	})
	return templateErr
}

func withTemplateLock(ctx context.Context, pool *pgxpool.Pool, fn func() error) error {
	const templateLockID int64 = 0x746573745f746d70
	if _, err := pool.Exec(ctx, `SELECT pg_advisory_lock($1)`, templateLockID); err != nil {
		return fmt.Errorf("acquire template lock: %w", err)
	}
	defer func() { _, _ = pool.Exec(ctx, `SELECT pg_advisory_unlock($1)`, templateLockID) }()
	return fn()
}

func buildTestTemplate(ctx context.Context, baseURL string, templateCfg config.Config) error {
	pool, err := EnsureAdminPool(ctx, baseURL)
	if err != nil {
		return fmt.Errorf("connect postgres for template: %w", err)
	}

	schemaSQL := pgx.Identifier{testTemplateSchema}.Sanitize()
	if _, err := pool.Exec(ctx, "DROP SCHEMA IF EXISTS "+schemaSQL+" CASCADE"); err != nil {
		return fmt.Errorf("drop stale template schema: %w", err)
	}
	if _, err := pool.Exec(ctx, "CREATE SCHEMA "+schemaSQL); err != nil {
		return fmt.Errorf("create template schema: %w", err)
	}

	templateURL := WithSearchPath(baseURL, testTemplateSchema)
	cfg := templateCfg
	cfg.DatabaseURL = templateURL
	cfg.LogDatabaseURL = templateURL
	cfg.LogSchemaIsolated = true
	cfg.StoreBootstrap.SchemaPrepared = false
	cfg.BootstrapMode = config.BootstrapDemo

	st, err := postgres.New(ctx, cfg)
	if err != nil {
		return fmt.Errorf("build test template: %w", err)
	}
	if pg, ok := st.(*postgres.Store); ok {
		if err := clearIngestRuntimeTables(ctx, pg); err != nil {
			pg.Close()
			return fmt.Errorf("clear template ingest tables: %w", err)
		}
		pg.Close()
	}
	return nil
}

func clearIngestRuntimeTables(ctx context.Context, st *postgres.Store) error {
	logPool := postgres.LogPool(st)
	_, err := logPool.Exec(ctx, `
		TRUNCATE logs, ingest_jobs RESTART IDENTITY;
		UPDATE reconcile_cursors SET last_log_id = 0, updated_at = NOW();
	`)
	return err
}

func readTemplateVersion(ctx context.Context, pool *pgxpool.Pool) (int, bool) {
	var comment *string
	err := pool.QueryRow(ctx, `
		SELECT obj_description(n.oid, 'pg_namespace')
		FROM pg_namespace n
		WHERE n.nspname = $1
	`, testTemplateSchema).Scan(&comment)
	if err != nil || comment == nil {
		return 0, false
	}
	var version int
	if _, err := fmt.Sscanf(*comment, "version:%d", &version); err != nil {
		return 0, false
	}
	return version, true
}

func markTemplateVersion(ctx context.Context, pool *pgxpool.Pool) error {
	schemaSQL := pgx.Identifier{testTemplateSchema}.Sanitize()
	_, err := pool.Exec(ctx, fmt.Sprintf("COMMENT ON SCHEMA %s IS 'version:%d'", schemaSQL, testTemplateVersion))
	return err
}

func OpenCloned(t *testing.T, baseURL string, templateCfg config.Config) Handle {
	t.Helper()
	if h, ok := CachedHandle(t); ok {
		return h
	}
	ctx := context.Background()
	if err := EnsureTemplate(ctx, baseURL, templateCfg); err != nil {
		t.Fatalf("ensure test template: %v", err)
	}
	pool, err := EnsureAdminPool(ctx, baseURL)
	if err != nil {
		t.Fatal(err)
	}
	schema := NewTestSchemaName()
	if err := CloneSchema(ctx, pool, testTemplateSchema, schema, clonePlan); err != nil {
		t.Fatalf("clone test schema: %v", err)
	}
	return registerTestSchema(t, pool, baseURL, schema)
}

func DropOrphanTestSchemas(ctx context.Context, baseURL string) error {
	pool, err := EnsureAdminPool(ctx, baseURL)
	if err != nil {
		return err
	}
	cleanupOrphanTestSchemas(ctx, pool)
	return nil
}
