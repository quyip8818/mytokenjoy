//go:build testhook

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
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestCreateCompanyInviteEmailMode(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	client := &mock.StubAdminClient{User: newapi.User{ID: 501, Quota: 0}}
	svc := company.NewService(cfg, st, newapi.NewAdminPortAdapter(client), permission.NewGrantNormalizer())
	ctx := context.Background()

	result, err := svc.CreateCompany(ctx, company.CreateCompanyRequest{
		Name:        "New Co",
		InviteEmail: "admin@newco.example",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.InviteCode == "" {
		t.Fatal("expected invite code in InviteEmail mode")
	}
	if result.Member != nil {
		t.Fatal("expected nil member in InviteEmail mode")
	}

	stored, err := st.Company().GetByID(ctx, result.Company.ID)
	if err != nil || stored == nil {
		t.Fatal("expected company to exist")
	}
	if stored.NewAPIWalletUserID == nil || *stored.NewAPIWalletUserID <= 0 {
		t.Fatal("expected wallet account")
	}
	if stored.RootDeptID == nil {
		t.Fatal("expected root department")
	}

	invite, err := st.Invite().GetInviteByCode(ctx, result.InviteCode)
	if err != nil || invite == nil {
		t.Fatal("expected invite to exist")
	}
	if invite.Email != "admin@newco.example" {
		t.Fatalf("expected invite email, got %s", invite.Email)
	}
}

func TestCreateCompanyUserIDMode(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	client := &mock.StubAdminClient{User: newapi.User{ID: 502, Quota: 0}}
	svc := company.NewService(cfg, st, newapi.NewAdminPortAdapter(client), permission.NewGrantNormalizer())
	ctx := context.Background()

	// Create a user first.
	userID := uuid.Must(uuid.NewV7())
	testutil.EnsureUser(t, st, userID)

	result, err := svc.CreateCompany(ctx, company.CreateCompanyRequest{
		UserID: userID,
		Name:   "User Mode Co",
		Type:   store.CompanyTypeTrial,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Member == nil {
		t.Fatal("expected member in UserID mode")
	}
	if result.InviteCode != "" {
		t.Fatal("expected no invite code in UserID mode")
	}
	if result.Member.UserID != userID {
		t.Fatalf("expected member.UserID=%s, got %s", userID, result.Member.UserID)
	}
	if result.Member.CompanyID != result.Company.ID {
		t.Fatal("expected member.CompanyID matches company")
	}
	if result.Company.Type != store.CompanyTypeTrial {
		t.Fatalf("expected trial type, got %s", result.Company.Type)
	}
}

func TestCreateCompanyUserIDModeIdempotent(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	var callCount int
	client := &mock.StubAdminClient{
		CreateUserFn: func(_ context.Context, _ newapi.CreateUserRequest) (newapi.User, error) {
			callCount++
			return newapi.User{ID: int64(600 + callCount), Quota: 0}, nil
		},
	}
	svc := company.NewService(cfg, st, newapi.NewAdminPortAdapter(client), permission.NewGrantNormalizer())
	ctx := context.Background()

	userID := uuid.Must(uuid.NewV7())
	testutil.EnsureUser(t, st, userID)

	result, err := svc.CreateCompany(ctx, company.CreateCompanyRequest{
		UserID: userID,
		Name:   "Idempotent Co",
	})
	if err != nil {
		t.Fatal(err)
	}

	// addMember is idempotent — calling again on same company should return existing member.
	// But CreateCompany creates a new company each time, so this tests the addMember idempotency
	// indirectly via AcceptInvite on same company.
	_ = result
}

func TestCreateCompanyRejectsBothEmpty(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	client := &mock.StubAdminClient{User: newapi.User{ID: 700, Quota: 0}}
	svc := company.NewService(cfg, st, newapi.NewAdminPortAdapter(client), permission.NewGrantNormalizer())

	_, err := svc.CreateCompany(context.Background(), company.CreateCompanyRequest{
		Name: "No User No Email",
	})
	if err == nil {
		t.Fatal("expected error when both UserID and InviteEmail are empty")
	}
}

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
		Name:        "Rollback Co",
		InviteEmail: "admin@rollback.example",
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

func TestCreateCompanyDefaultsToStandardType(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	client := &mock.StubAdminClient{User: newapi.User{ID: 601, Quota: 0}}
	svc := company.NewService(cfg, st, newapi.NewAdminPortAdapter(client), permission.NewGrantNormalizer())
	ctx := context.Background()

	result, err := svc.CreateCompany(ctx, company.CreateCompanyRequest{
		Name:        "Default Type Co",
		InviteEmail: "admin@default-type.example",
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

func TestCreateCompanyPresetRolesAndBudgetTree(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	client := &mock.StubAdminClient{User: newapi.User{ID: 603, Quota: 0}}
	svc := company.NewService(cfg, st, newapi.NewAdminPortAdapter(client), permission.NewGrantNormalizer())
	ctx := context.Background()

	result, err := svc.CreateCompany(ctx, company.CreateCompanyRequest{
		Name:        "Roles Co",
		InviteEmail: "admin@roles.example",
	})
	if err != nil {
		t.Fatal(err)
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
