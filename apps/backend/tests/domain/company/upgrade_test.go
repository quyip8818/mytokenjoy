//go:build testhook

package company_test

import (
	"testing"

	"github.com/google/uuid"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

// spyCacheInvalidator records calls to InvalidateCompany.
type spyCacheInvalidator struct {
	types.NoopPrecheckCacheInvalidator
	invalidatedCompanies []uuid.UUID
}

func (s *spyCacheInvalidator) InvalidateCompany(id uuid.UUID) {
	s.invalidatedCompanies = append(s.invalidatedCompanies, id)
}

func TestUpgradeToStandardChangesTypeAndInvalidatesCache(t *testing.T) {
	t.Parallel()
	companyID := uuid.MustParse("00000000-0000-7000-0000-000000009301")
	_, st := testutil.NewTestStore(t)
	ctx := testutil.CtxForCompany(companyID)

	// Create a trial company.
	if err := st.Company().Create(ctx, store.Company{
		ID: companyID, Name: "Upgrade Test", Type: store.CompanyTypeTrial, Status: store.CompanyStatusActive,
	}); err != nil {
		t.Fatal(err)
	}

	// Seed trial credit.
	if err := domainbilling.SeedTrialCredit(ctx, st, companyID, 10000, nil); err != nil {
		t.Fatal(err)
	}

	// Wire service with spy cache invalidator.
	spy := &spyCacheInvalidator{}
	cfg := testutil.TestConfig()
	svc := domaincompany.NewService(cfg, st, &mock.StubAdminClient{}, permission.NewGrantNormalizer(),
		domaincompany.WithCompanyCacheInvalidator(spy))

	// Upgrade.
	if err := svc.UpgradeToStandard(ctx, companyID); err != nil {
		t.Fatal(err)
	}

	// Verify type changed to standard.
	co, err := st.Company().GetByID(ctx, companyID)
	if err != nil || co == nil {
		t.Fatal("expected company after upgrade")
	}
	if co.Type != store.CompanyTypeStandard {
		t.Fatalf("type after upgrade: got %q want %q", co.Type, store.CompanyTypeStandard)
	}

	// Verify wallet zeroed (mock lots expired).
	if co.WalletQuotaRemain != 0 {
		t.Fatalf("wallet after upgrade: got %v want 0", co.WalletQuotaRemain)
	}

	// Verify cache was invalidated.
	if len(spy.invalidatedCompanies) != 1 || spy.invalidatedCompanies[0] != companyID {
		t.Fatalf("expected cache invalidation for %s, got %v", companyID, spy.invalidatedCompanies)
	}
}

func TestUpgradeToStandardRejectsStandardCompany(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()

	// Default seed company is standard — upgrade should fail.
	svc := domaincompany.NewService(cfg, st, &mock.StubAdminClient{}, permission.NewGrantNormalizer())
	err := svc.UpgradeToStandard(ctx, uuid.MustParse("00000000-0000-7000-8000-000000000001"))
	if err == nil {
		t.Fatal("expected error upgrading standard company")
	}
}

func TestUpgradeToStandardFromDemo(t *testing.T) {
	t.Parallel()
	companyID := uuid.MustParse("00000000-0000-7000-0000-000000009302")
	_, st := testutil.NewTestStore(t)
	ctx := testutil.CtxForCompany(companyID)

	// Create a demo company.
	if err := st.Company().Create(ctx, store.Company{
		ID: companyID, Name: "Demo Upgrade", Type: store.CompanyTypeDemo, Status: store.CompanyStatusActive,
	}); err != nil {
		t.Fatal(err)
	}

	// Seed trial credit.
	if err := domainbilling.SeedTrialCredit(ctx, st, companyID, 5000, nil); err != nil {
		t.Fatal(err)
	}

	cfg := testutil.TestConfig()
	svc := domaincompany.NewService(cfg, st, &mock.StubAdminClient{}, permission.NewGrantNormalizer())

	if err := svc.UpgradeToStandard(ctx, companyID); err != nil {
		t.Fatal(err)
	}

	co, err := st.Company().GetByID(ctx, companyID)
	if err != nil || co == nil {
		t.Fatal("expected company after demo upgrade")
	}
	if co.Type != store.CompanyTypeStandard {
		t.Fatalf("type after demo upgrade: got %q want %q", co.Type, store.CompanyTypeStandard)
	}
	if co.WalletQuotaRemain != 0 {
		t.Fatalf("wallet after demo upgrade: got %v want 0", co.WalletQuotaRemain)
	}
}
