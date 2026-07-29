//go:build testhook

package pg

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/store/postgres"
)

const (
	templateDBName      = "test_template_db"
	testTemplateVersion = 47 // bump when schema/seed changes
)

var (
	templateOnce sync.Once
	templateErr  error
)

// EnsureTemplateDB creates (or verifies) the template database used by
// CREATE DATABASE ... TEMPLATE in OpenCloned. Safe for concurrent callers.
func EnsureTemplateDB(ctx context.Context, baseURL string, templateCfg config.Config) error {
	templateOnce.Do(func() {
		templateErr = buildOrVerifyTemplateDB(ctx, baseURL, templateCfg)
	})
	return templateErr
}

// templateLockID is a fixed advisory lock key used to serialize template DB
// creation across parallel test processes (go test -p N).
const templateLockID = 987654321

func buildOrVerifyTemplateDB(ctx context.Context, baseURL string, templateCfg config.Config) error {
	adminConn, err := pgx.Connect(ctx, baseURL)
	if err != nil {
		return fmt.Errorf("connect admin: %w", err)
	}
	defer adminConn.Close(ctx)

	// Acquire a Postgres-level advisory lock so that parallel test processes
	// (separate OS processes from go test -p 8) serialize template DB creation.
	// The lock is automatically released when the connection closes.
	if _, err := adminConn.Exec(ctx, "SELECT pg_advisory_lock($1)", templateLockID); err != nil {
		return fmt.Errorf("acquire advisory lock: %w", err)
	}

	// Clean up orphan test databases left by previous abnormal exits.
	cleanupOrphanTestDatabases(ctx, adminConn)

	// Re-check version under lock — another process may have finished first.
	if version, ok := readDBVersion(ctx, adminConn, templateDBName); ok && version == testTemplateVersion {
		return nil
	}

	// Terminate any lingering connections to the template DB.
	terminateDBConnections(ctx, adminConn, templateDBName)

	// Drop stale template DB.
	_, _ = adminConn.Exec(ctx, fmt.Sprintf(
		"DROP DATABASE IF EXISTS %s",
		pgx.Identifier{templateDBName}.Sanitize(),
	))

	// Create fresh template DB.
	_, err = adminConn.Exec(ctx, fmt.Sprintf(
		"CREATE DATABASE %s",
		pgx.Identifier{templateDBName}.Sanitize(),
	))
	if err != nil {
		return fmt.Errorf("create template db: %w", err)
	}

	// Bootstrap schema + seed inside the template DB.
	templateURL := replaceDBName(baseURL, templateDBName)
	cfg := templateCfg
	cfg.DatabaseURL = templateURL
	cfg.LogDatabaseURL = templateURL
	cfg.StoreBootstrap.SkipSchema = false
	cfg.StoreBootstrap.SkipSeed = false

	st, err := postgres.New(ctx, cfg)
	if err != nil {
		return fmt.Errorf("bootstrap template db: %w", err)
	}
	if pg, ok := st.(*postgres.Store); ok {
		pg.Close()
	}

	// Stamp version so subsequent runs skip rebuild.
	return markDBVersion(ctx, adminConn, templateDBName, testTemplateVersion)
}

func terminateDBConnections(ctx context.Context, conn *pgx.Conn, dbName string) {
	_, _ = conn.Exec(ctx, `
		SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = $1 AND pid <> pg_backend_pid()
	`, dbName)
}

func cleanupOrphanTestDatabases(ctx context.Context, conn *pgx.Conn) {
	rows, _ := conn.Query(ctx, `
		SELECT datname FROM pg_database
		WHERE datname ~ '^test_[0-9a-f]{16}$'
		AND datname <> $1
	`, templateDBName)
	if rows == nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		terminateDBConnections(ctx, conn, name)
		_, _ = conn.Exec(ctx, "DROP DATABASE IF EXISTS "+pgx.Identifier{name}.Sanitize())
	}
}
