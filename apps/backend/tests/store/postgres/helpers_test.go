//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/tests/testutil"
)

func requireDatabaseURL(t *testing.T) string {
	t.Helper()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set")
	}
	return dbURL
}

func testDBPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), requireDatabaseURL(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pool.Close() })
	return pool
}

func reopenPostgresStore(t *testing.T, dbURL string) store.Store {
	t.Helper()
	cfg := testutil.TestConfig()
	cfg.DatabaseURL = dbURL
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
	return reopenPostgresStore(t, requireDatabaseURL(t))
}
