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
)

const testTemplateSchema = "test_template"

type Handle struct {
	BaseURL string
	Schema  string
	URL     string
}

// schemaByTest keys on *testing.T (not t.Name()) so -count with t.Parallel()
// does not reuse one schema across concurrent invocations of the same test name.
var schemaByTest sync.Map

func CachedHandle(t *testing.T) (Handle, bool) {
	v, ok := schemaByTest.Load(t)
	if !ok {
		return Handle{}, false
	}
	return *v.(*Handle), true
}

func NewTestSchemaName() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return "test_" + hex.EncodeToString(b[:])
}

func WithSearchPath(dbURL, schema string) string {
	u, err := url.Parse(dbURL)
	if err != nil {
		panic(fmt.Sprintf("parse database url: %v", err))
	}
	q := u.Query()
	q.Set("options", fmt.Sprintf("-c search_path=%s,public", schema))
	u.RawQuery = q.Encode()
	return u.String()
}

func registerTestSchema(t *testing.T, pool *pgxpool.Pool, baseURL, schema string) Handle {
	t.Helper()
	h := Handle{
		BaseURL: baseURL,
		Schema:  schema,
		URL:     WithSearchPath(baseURL, schema),
	}
	schemaByTest.Store(t, &h)
	schemaSQL := pgx.Identifier{schema}.Sanitize()
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DROP SCHEMA IF EXISTS "+schemaSQL+" CASCADE")
		schemaByTest.Delete(t)
	})
	return h
}

func OpenSlow(t *testing.T, baseURL string) Handle {
	t.Helper()
	if h, ok := CachedHandle(t); ok {
		return h
	}
	pool, err := EnsureAdminPool(context.Background(), baseURL)
	if err != nil {
		t.Fatal(err)
	}
	schema := NewTestSchemaName()
	schemaSQL := pgx.Identifier{schema}.Sanitize()
	if _, err := pool.Exec(context.Background(), "CREATE SCHEMA "+schemaSQL); err != nil {
		t.Fatal(err)
	}
	return registerTestSchema(t, pool, baseURL, schema)
}

func cleanupOrphanTestSchemas(ctx context.Context, pool *pgxpool.Pool) {
	rows, err := pool.Query(ctx, `
		SELECT nspname
		FROM pg_namespace
		WHERE nspname ~ '^test_[0-9a-f]{16}$'
		ORDER BY oid
	`)
	if err != nil {
		return
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return
		}
		schemas = append(schemas, name)
	}
	if err := rows.Err(); err != nil {
		return
	}
	const keepRecent = 32
	if len(schemas) <= keepRecent {
		return
	}
	for _, schema := range schemas[:len(schemas)-keepRecent] {
		schemaSQL := pgx.Identifier{schema}.Sanitize()
		_, _ = pool.Exec(ctx, "DROP SCHEMA IF EXISTS "+schemaSQL+" CASCADE")
	}
}
