package company_test

import (
	"testing"

	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestGateSuspendedCompany(t *testing.T) {
	t.Parallel()
	cfg := testutil.TestConfig()
	gate := domaincompany.NewGate(cfg)
	ctx := domaincompany.WithContext(testutil.Ctx(), domaincompany.Context{
		CompanyID: 1, Status: store.CompanyStatusSuspended,
	})
	if !gate.IsSuspended(ctx) {
		t.Fatal("expected suspended company")
	}
}

func TestGateActiveCompany(t *testing.T) {
	t.Parallel()
	cfg := testutil.TestConfig()
	gate := domaincompany.NewGate(cfg)
	ctx := domaincompany.WithContext(testutil.Ctx(), domaincompany.Context{
		CompanyID: 1, Status: store.CompanyStatusActive,
	})
	if gate.IsSuspended(ctx) {
		t.Fatal("expected active company not suspended")
	}
}
