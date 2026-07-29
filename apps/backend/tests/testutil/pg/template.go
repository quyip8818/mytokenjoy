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

const testTemplateVersion = 49 // bump when schema/seed changes — 两个 template 共享版本号

// errOnce supports retry on failure (only marks done on success).
type errOnce struct {
	mu   sync.Mutex
	done bool
	err  error
}

func (o *errOnce) Do(f func() error) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.done {
		return o.err
	}
	o.err = f()
	if o.err == nil {
		o.done = true
	}
	return o.err
}

var templateOnces sync.Map // map[string]*errOnce — keyed by mode

// EnsureTemplateDB creates (or verifies) the template database for the given mode.
// mode is "saas" or "local". Safe for concurrent callers.
func EnsureTemplateDB(ctx context.Context, baseURL string, mode string, templateCfg config.Config) error {
	v, _ := templateOnces.LoadOrStore(mode, &errOnce{})
	return v.(*errOnce).Do(func() error {
		return buildOrVerifyTemplateDB(ctx, baseURL, mode, templateCfg)
	})
}

// TemplateDBName returns the template database name for a given mode.
func TemplateDBName(mode string) string {
	return "template_" + mode // template_saas / template_local
}

func advisoryLockID(mode string) int64 {
	if mode == "local" {
		return 987654322
	}
	return 987654321
}

func buildOrVerifyTemplateDB(ctx context.Context, baseURL string, mode string, templateCfg config.Config) error {
	dbName := TemplateDBName(mode)
	lockID := advisoryLockID(mode)

	adminConn, err := pgx.Connect(ctx, baseURL)
	if err != nil {
		return fmt.Errorf("connect admin: %w", err)
	}
	defer adminConn.Close(ctx)

	// Acquire a Postgres-level advisory lock so that parallel test processes
	// (separate OS processes from go test -p N) serialize template DB creation.
	// The lock is automatically released when the connection closes.
	if _, err := adminConn.Exec(ctx, "SELECT pg_advisory_lock($1)", lockID); err != nil {
		return fmt.Errorf("acquire advisory lock: %w", err)
	}

	// Clean up orphan test databases left by previous abnormal exits.
	cleanupOrphanTestDatabases(ctx, adminConn)

	// Re-check version under lock — another process may have finished first.
	if version, ok := readDBVersion(ctx, adminConn, dbName); ok && version == testTemplateVersion {
		return nil
	}

	// Terminate any lingering connections to the template DB.
	terminateDBConnections(ctx, adminConn, dbName)

	// Drop stale template DB.
	_, _ = adminConn.Exec(ctx, fmt.Sprintf(
		"DROP DATABASE IF EXISTS %s",
		pgx.Identifier{dbName}.Sanitize(),
	))

	// Create fresh template DB.
	_, err = adminConn.Exec(ctx, fmt.Sprintf(
		"CREATE DATABASE %s",
		pgx.Identifier{dbName}.Sanitize(),
	))
	if err != nil {
		return fmt.Errorf("create template db: %w", err)
	}

	// Bootstrap schema + seed inside the template DB.
	templateURL := replaceDBName(baseURL, dbName)
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
	return markDBVersion(ctx, adminConn, dbName, testTemplateVersion)
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
	`)
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
