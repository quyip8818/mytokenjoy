package company_test

import (
	"context"
	"errors"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store/memory"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestCreateCompanyRollsBackOnCreateUserFailure(t *testing.T) {
	cfg := testutil.TestConfig(testutil.WithNewAPIEnabled(true))
	st := memory.New(seed.Load(cfg))
	client := &mock.StubAdminClient{
		CreateUserFn: func(context.Context, newapi.CreateUserRequest) (newapi.User, error) {
			return newapi.User{}, errors.New("newapi unavailable")
		},
	}
	svc := company.NewService(cfg, st, client)
	ctx := context.Background()

	before, err := st.Company().List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	beforeCount := len(before)

	_, err = svc.CreateCompany(ctx, company.CreateCompanyRequest{
		Slug:            "rollback-co",
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
	cfg := testutil.TestConfig(testutil.WithNewAPIEnabled(true))
	st := memory.New(seed.Load(cfg))
	client := &mock.StubAdminClient{
		User: newapi.User{ID: 501, Quota: 0},
	}
	svc := company.NewService(cfg, st, client)
	ctx := context.Background()

	result, err := svc.CreateCompany(ctx, company.CreateCompanyRequest{
		Slug:            "new-co",
		Name:            "New Co",
		SuperAdminEmail: "admin@newco.example",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.InviteToken == "" {
		t.Fatal("expected invite token")
	}

	stored, err := st.Company().GetByID(ctx, result.Company.ID)
	if err != nil || stored == nil {
		t.Fatal("expected company to exist")
	}
	if stored.NewAPIWalletUserID == nil || *stored.NewAPIWalletUserID != 501 {
		t.Fatalf("expected wallet account 501, got %v", stored.NewAPIWalletUserID)
	}
	invite, err := st.Invite().GetInviteByToken(ctx, result.InviteToken)
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
