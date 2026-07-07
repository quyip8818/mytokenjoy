package postgres_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/tests/testutil"
)

func testPostgresStore(t *testing.T) store.Store {
	t.Helper()
	_, st := testutil.NewTestStore(t)
	return st
}

func reopenPostgresStore(t *testing.T, dbURL string) store.Store {
	t.Helper()
	schemaURL := testutil.TestSchemaURL(t)
	if dbURL != "" && dbURL != schemaURL {
		t.Fatalf("unexpected database url for test schema")
	}
	cfg := testutil.TestConfig()
	cfg.DatabaseURL = schemaURL
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

func testDBPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), testutil.TestSchemaURL(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pool.Close() })
	return pool
}
