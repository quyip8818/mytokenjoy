package company_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestResolveFromMember(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := domaincompany.NewService(cfg, st, newapi.NewAdminPortAdapter(&mock.StubAdminClient{}), permission.NewGrantNormalizer())
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
	svc := domaincompany.NewService(cfg, st, newapi.NewAdminPortAdapter(&mock.StubAdminClient{}), permission.NewGrantNormalizer())
	ctx := testutil.Ctx()

	got, err := svc.ResolveCompanyContext(ctx, contract.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if got.CompanyID != contract.DefaultCompanyID {
		t.Fatalf("expected company %d, got %d", contract.DefaultCompanyID, got.CompanyID)
	}
}

func TestResolveCompanyContext_MissingIsNotFound(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := domaincompany.NewService(cfg, st, newapi.NewAdminPortAdapter(&mock.StubAdminClient{}), permission.NewGrantNormalizer())
	ctx := testutil.Ctx()

	_, err := svc.ResolveCompanyContext(ctx, uuid.MustParse("00000000-0000-7000-0000-3b9ac9ff0000"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !domain.IsNotFound(err) {
		t.Fatalf("expected NotFound, got %v", err)
	}
}
