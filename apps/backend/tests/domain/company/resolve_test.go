package company_test

import (
	"testing"

	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestResolveFromMember(t *testing.T) {
	cfg, st := testutil.NewTestStore(t)
	svc := domaincompany.NewService(cfg, st, &mock.StubAdminClient{})
	ctx := testutil.Ctx()

	got, err := svc.ResolveFromMember(ctx, seed.IDMember1)
	if err != nil {
		t.Fatal(err)
	}
	if got.CompanyID != seed.DefaultCompanyID {
		t.Fatalf("expected company %d, got %d", seed.DefaultCompanyID, got.CompanyID)
	}
}

func TestResolveCompanyContext(t *testing.T) {
	cfg, st := testutil.NewTestStore(t)
	svc := domaincompany.NewService(cfg, st, &mock.StubAdminClient{})
	ctx := testutil.Ctx()

	got, err := svc.ResolveCompanyContext(ctx, seed.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if got.CompanyID != seed.DefaultCompanyID {
		t.Fatalf("expected company %d, got %d", seed.DefaultCompanyID, got.CompanyID)
	}
}
