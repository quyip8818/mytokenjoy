package postgres

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
)

func TestEnsureBootstrapCompanyRejectsSaaSRangeWhenSupportSaasDisabled(t *testing.T) {
	ctx := context.Background()
	schemaURL := testSchemaURL(t)
	cfg := config.Config{
		DatabaseURL:       schemaURL,
		SupportSaas:       false,
		CompanyName:       "Demo Company",
		TokenJoyCompanyID: contract.TokenJoyCompanyID,
		LocalCompanyID:    contract.LocalCompanyID,
		DefaultCompanyID:  contract.LocalCompanyID,
	}

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("create postgres pool: %v", err)
	}
	defer pool.Close()

	if err := applySchema(ctx, pool); err != nil {
		t.Fatalf("apply schema: %v", err)
	}

	// Seed a SaaS-range company id. When SUPPORT_SAAS=false, bootstrap validation
	// must reject the startup due to existing SaaS-range company rows.
	if _, err := pool.Exec(ctx, `
		INSERT INTO companies (id, slug, name, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING
	`, saasMinCompanyID, "invalid-saas-range", "Invalid SaaS Range", store.CompanyStatusActive); err != nil {
		t.Fatalf("seed invalid company: %v", err)
	}

	err = ensureBootstrapCompany(ctx, pool, cfg)
	if err == nil {
		t.Fatal("expected ensureBootstrapCompany to fail")
	}
	if !strings.Contains(err.Error(), "SUPPORT_SAAS=false but found SaaS-range company ids") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func testSchemaURL(t *testing.T) string {
	t.Helper()

	baseURL := os.Getenv("DATABASE_URL")
	if baseURL == "" {
		baseURL = config.DefaultDatabaseURL
	}
	if baseURL == "" {
		t.Fatal("DATABASE_URL required; run: pnpm start:postgres")
	}

	schema := newTestSchemaName()
	adminPool, err := pgxpool.New(context.Background(), baseURL)
	if err != nil {
		t.Fatalf("create postgres admin pool: %v", err)
	}

	schemaSQL := pgx.Identifier{schema}.Sanitize()
	if _, err := adminPool.Exec(context.Background(), "CREATE SCHEMA "+schemaSQL); err != nil {
		adminPool.Close()
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, _ = adminPool.Exec(context.Background(), "DROP SCHEMA IF EXISTS "+schemaSQL+" CASCADE")
		adminPool.Close()
	})

	return withSearchPath(baseURL, schema)
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
