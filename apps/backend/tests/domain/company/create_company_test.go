package company_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestCreateCompanyRollsBackOnCreateUserFailure(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	client := &mock.StubAdminClient{
		CreateUserFn: func(context.Context, newapi.CreateUserRequest) (newapi.User, error) {
			return newapi.User{}, errors.New("newapi unavailable")
		},
	}
	svc := company.NewService(cfg, st, newapi.NewAdminPortAdapter(client), permission.NewGrantNormalizer())
	ctx := context.Background()

	before, err := st.Company().List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	beforeCount := len(before)

	_, err = svc.CreateCompany(ctx, company.CreateCompanyRequest{
		Name:            "Rollback Co",
		SuperAdminEmail: "admin@rollback.example",
	})
	if err == nil {
		t.Fatal("expected create company to fail")
	}

	after, err := st.Company().List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != beforeCount {
		t.Fatalf("expected company count unchanged, before=%d after=%d", beforeCount, len(after))
	}
}

func TestCreateCompanyPersistsWalletAndInvite(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	client := &mock.StubAdminClient{
		User: newapi.User{ID: 501, Quota: 0},
	}
	svc := company.NewService(cfg, st, newapi.NewAdminPortAdapter(client), permission.NewGrantNormalizer())
	ctx := context.Background()

	result, err := svc.CreateCompany(ctx, company.CreateCompanyRequest{
		Name:            "New Co",
		SuperAdminEmail: "admin@newco.example",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.InviteCode == "" {
		t.Fatal("expected invite token")
	}

	stored, err := st.Company().GetByID(ctx, result.Company.ID)
	if err != nil || stored == nil {
		t.Fatal("expected company to exist")
	}
	if stored.NewAPIWalletUserID == nil || *stored.NewAPIWalletUserID != 501 {
		t.Fatalf("expected wallet account 501, got %v", stored.NewAPIWalletUserID)
	}
	tbs, err := st.TenantBackgroundState().Get(ctx, result.Company.ID)
	if err != nil {
		t.Fatal(err)
	}
	if tbs == nil {
		t.Fatal("expected tenant_background_state row")
	}
	invite, err := st.Invite().GetInviteByCode(ctx, result.InviteCode)
	if err != nil || invite == nil {
		t.Fatal("expected invite to exist")
	}
	if invite.CompanyID != result.Company.ID {
		t.Fatalf("expected invite company %d, got %d", result.Company.ID, invite.CompanyID)
	}
	if stored.RootDeptID == nil {
		t.Fatal("expected root department")
	}
	companyCtx := company.WithContext(ctx, company.Context{CompanyID: result.Company.ID})
	tree, err := common.LoadBudgetTree(companyCtx, st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	if len(tree) == 0 {
		t.Fatal("expected budget tree root")
	}
	roles, err := st.Org().Roles(companyCtx)
	if err != nil {
		t.Fatal(err)
	}
	if len(roles) == 0 {
		t.Fatal("expected preset roles for new company")
	}
}

func TestCreateCompanyAllocatesNextID(t *testing.T) {
	cfg, st := testutil.NewTestStore(t)
	client := &mock.StubAdminClient{User: newapi.User{ID: 503, Quota: 0}}
	svc := company.NewService(cfg, st, newapi.NewAdminPortAdapter(client), permission.NewGrantNormalizer())
	ctx := context.Background()

	result, err := svc.CreateCompany(ctx, company.CreateCompanyRequest{
		Name:            "Next ID",
		SuperAdminEmail: "admin@next-id.example",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Company.ID == uuid.Nil {
		t.Fatal("expected non-nil company id")
	}
}

func TestCreateCompanyDefaultsToStandardType(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	client := &mock.StubAdminClient{User: newapi.User{ID: 601, Quota: 0}}
	svc := company.NewService(cfg, st, newapi.NewAdminPortAdapter(client), permission.NewGrantNormalizer())
	ctx := context.Background()

	result, err := svc.CreateCompany(ctx, company.CreateCompanyRequest{
		Name:            "Default Type Co",
		SuperAdminEmail: "admin@default-type.example",
	})
	if err != nil {
		t.Fatal(err)
	}
	stored, err := st.Company().GetByID(ctx, result.Company.ID)
	if err != nil || stored == nil {
		t.Fatal("expected company")
	}
	if stored.Type != "standard" {
		t.Fatalf("expected type=standard, got %s", stored.Type)
	}
}

func TestCreateCompanyRespectsExplicitType(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	client := &mock.StubAdminClient{User: newapi.User{ID: 602, Quota: 0}}
	svc := company.NewService(cfg, st, newapi.NewAdminPortAdapter(client), permission.NewGrantNormalizer())
	ctx := context.Background()

	result, err := svc.CreateCompany(ctx, company.CreateCompanyRequest{
		Name:            "Demo Co",
		Type:            "demo",
		SuperAdminEmail: "admin@demo.example",
	})
	if err != nil {
		t.Fatal(err)
	}
	stored, err := st.Company().GetByID(ctx, result.Company.ID)
	if err != nil || stored == nil {
		t.Fatal("expected company")
	}
	if stored.Type != "demo" {
		t.Fatalf("expected type=demo, got %s", stored.Type)
	}
}
