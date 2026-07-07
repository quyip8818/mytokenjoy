//go:build integration

package postgres_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/tests/testutil"
)

type testDB struct {
	baseURL string
	schema  string
	url     string
}

var testDBByName sync.Map

func requireDatabaseURL(t *testing.T) string {
	t.Helper()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set")
	}
	return dbURL
}

func getTestDB(t *testing.T) *testDB {
	t.Helper()
	if v, ok := testDBByName.Load(t.Name()); ok {
		return v.(*testDB)
	}
	baseURL := requireDatabaseURL(t)
	schema := newTestSchemaName()
	adminPool, err := pgxpool.New(context.Background(), baseURL)
	if err != nil {
		t.Fatal(err)
	}
	schemaSQL := pgx.Identifier{schema}.Sanitize()
	if _, err := adminPool.Exec(context.Background(), "CREATE SCHEMA "+schemaSQL); err != nil {
		adminPool.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = adminPool.Exec(context.Background(), "DROP SCHEMA IF EXISTS "+schemaSQL+" CASCADE")
		adminPool.Close()
		testDBByName.Delete(t.Name())
	})
	h := &testDB{
		baseURL: baseURL,
		schema:  schema,
		url:     withSearchPath(baseURL, schema),
	}
	testDBByName.Store(t.Name(), h)
	return h
}

func newTestSchemaName() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return "test_" + hex.EncodeToString(b[:])
}

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

func testDBPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), getTestDB(t).url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pool.Close() })
	return pool
}

func reopenPostgresStore(t *testing.T, dbURL string) store.Store {
	t.Helper()
	h := getTestDB(t)
	if dbURL != h.baseURL && dbURL != h.url {
		t.Fatalf("unexpected database url for test schema")
	}
	cfg := testutil.TestConfig()
	cfg.DatabaseURL = h.url
	st, err := postgres.New(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if pg, ok := st.(*postgres.Store); ok {
			pg.Close()
		}
	})
	return st
}

func testPostgresStore(t *testing.T) store.Store {
	t.Helper()
	return reopenPostgresStore(t, getTestDB(t).url)
}
