//go:build testhook

package pg

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/config"
)

// Handle is the per-test database reference returned by OpenCloned or OpenSlow.
type Handle struct {
	DBName string
	URL    string
}

// handleByTest keys on *testing.T so -count with t.Parallel() does not reuse
// one database across concurrent invocations of the same test name.
var handleByTest sync.Map

func CachedHandle(t *testing.T) (Handle, bool) {
	v, ok := handleByTest.Load(t)
	if !ok {
		return Handle{}, false
	}
	return *v.(*Handle), true
}

// OpenCloned creates a fresh database by cloning the template DB via
// CREATE DATABASE ... TEMPLATE. Each test gets its own isolated database.
func OpenCloned(t *testing.T, baseURL string, templateCfg config.Config) Handle {
	t.Helper()
	if h, ok := CachedHandle(t); ok {
		return h
	}

	ctx := context.Background()
	if err := EnsureTemplateDB(ctx, baseURL, templateCfg); err != nil {
		t.Fatalf("ensure template db: %v", err)
	}

	dbName := newTestDBName()

	// CREATE DATABASE cannot run inside a transaction; use a dedicated admin connection.
	adminConn, err := pgx.Connect(ctx, baseURL)
	if err != nil {
		t.Fatalf("admin connect: %v", err)
	}

	// Defensive: terminate any lingering connections to the template DB.
	// EnsureTemplateDB closes its bootstrap connection, but this guards against
	// edge cases like idle pool keepalive from a previous failed run.
	terminateDBConnections(ctx, adminConn, templateDBName)

	_, err = adminConn.Exec(ctx, fmt.Sprintf(
		"CREATE DATABASE %s TEMPLATE %s",
		pgx.Identifier{dbName}.Sanitize(),
		pgx.Identifier{templateDBName}.Sanitize(),
	))
	adminConn.Close(ctx)
	if err != nil {
		t.Fatalf("clone database: %v", err)
	}

	h := Handle{
		DBName: dbName,
		URL:    replaceDBName(baseURL, dbName),
	}
	handleByTest.Store(t, &h)

	t.Cleanup(func() {
		conn, err := pgx.Connect(context.Background(), baseURL)
		if err != nil {
			return
		}
		defer conn.Close(context.Background())
		terminateDBConnections(context.Background(), conn, dbName)
		_, _ = conn.Exec(context.Background(), fmt.Sprintf(
			"DROP DATABASE IF EXISTS %s",
			pgx.Identifier{dbName}.Sanitize(),
		))
		handleByTest.Delete(t)
	})
	return h
}

// OpenSlow creates an empty schema in the base database for minimal-bootstrap tests.
// No template cloning — the caller is expected to run its own bootstrap.
func OpenSlow(t *testing.T, baseURL string) Handle {
	t.Helper()
	if h, ok := CachedHandle(t); ok {
		return h
	}
	ctx := context.Background()
	pool, err := EnsureAdminPool(ctx, baseURL)
	if err != nil {
		t.Fatal(err)
	}
	schema := newTestSchemaName()
	schemaSQL := pgx.Identifier{schema}.Sanitize()
	if _, err := pool.Exec(ctx, "CREATE SCHEMA "+schemaSQL); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	h := Handle{
		DBName: schema,
		URL:    withSearchPath(baseURL, schema),
	}
	handleByTest.Store(t, &h)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DROP SCHEMA IF EXISTS "+schemaSQL+" CASCADE")
		handleByTest.Delete(t)
	})
	return h
}

func newTestDBName() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return "test_" + hex.EncodeToString(b[:])
}

func newTestSchemaName() string {
	return newTestDBName() // same format, reuse
}

func replaceDBName(connStr, newDB string) string {
	u, err := url.Parse(connStr)
	if err != nil {
		panic(err)
	}
	u.Path = "/" + newDB
	return u.String()
}

// withSearchPath is used only by OpenSlow for schema-level isolation.
func withSearchPath(dbURL, schema string) string {
	u, err := url.Parse(dbURL)
	if err != nil {
		panic(fmt.Sprintf("parse database url: %v", err))
	}
	q := u.Query()
	q.Set("options", fmt.Sprintf("-c search_path=%s,public", schema))
	u.RawQuery = q.Encode()
	return u.String()
}

// DropOrphanTestDatabases cleans up orphan test databases left by abnormal exits.
// Called from cmd/testdbclean.
func DropOrphanTestDatabases(ctx context.Context, baseURL string) error {
	conn, err := pgx.Connect(ctx, baseURL)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	cleanupOrphanTestDatabases(ctx, conn)
	return nil
}

// EnsureAdminPool returns a shared pool for admin operations (schema DDL in OpenSlow).
func EnsureAdminPool(ctx context.Context, baseURL string) (*pgxpool.Pool, error) {
	adminOnce.Do(func() {
		adminPool, adminErr = pgxpool.New(ctx, baseURL)
	})
	return adminPool, adminErr
}

var (
	adminOnce sync.Once
	adminPool *pgxpool.Pool
	adminErr  error
)
