package company_test

import (
	"testing"

	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestResolveFromMember(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := domaincompany.NewService(cfg, st, &mock.StubAdminClient{})
	ctx := testutil.Ctx()

	got, err := svc.ResolveFromMember(ctx, contract.IDMember1)
	if err != nil {
		t.Fatal(err)
	}
	if got.CompanyID != contract.DefaultCompanyID {
		t.Fatalf("expected company %d, got %d", contract.DefaultCompanyID, got.CompanyID)
	}
}

func TestResolveCompanyContext(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := domaincompany.NewService(cfg, st, &mock.StubAdminClient{})
	ctx := testutil.Ctx()

	got, err := svc.ResolveCompanyContext(ctx, contract.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if got.CompanyID != contract.DefaultCompanyID {
		t.Fatalf("expected company %d, got %d", contract.DefaultCompanyID, got.CompanyID)
	}
}
