package postgres_test

import (
	"context"
	"strings"
	"testing"

	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestEnsureBootstrapCompanyRejectsSaaSRangeWhenSupportSaasDisabled(t *testing.T) {
	ctx := context.Background()
	pool := testDBPool(t)
	cfg := testutil.TestConfig()
	cfg.SupportSaas = false
	cfg.CompanyName = "Demo Company"
	cfg.TokenJoyCompanyID = 1
	cfg.LocalCompanyID = 2

	if _, err := pool.Exec(ctx, `
		INSERT INTO companies (id, slug, name, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING
	`, postgres.SaaSMinCompanyIDForTest, "invalid-saas-range", "Invalid SaaS Range", store.CompanyStatusActive); err != nil {
		t.Fatalf("seed invalid company: %v", err)
	}

	err := postgres.EnsureBootstrapCompanyForTest(ctx, pool, cfg)
	if err == nil {
		t.Fatal("expected ensureBootstrapCompany to fail")
	}
	if !strings.Contains(err.Error(), "SUPPORT_SAAS=false but found SaaS-range company ids") {
		t.Fatalf("unexpected error: %v", err)
	}
}
