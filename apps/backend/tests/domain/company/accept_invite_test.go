//go:build testhook

package company_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func TestAcceptInviteCreatesMembers(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	client := &mock.StubAdminClient{User: newapi.User{ID: 700, Quota: 0}}
	svc := company.NewService(cfg, st, newapi.NewAdminPortAdapter(client), permission.NewGrantNormalizer())
	ctx := context.Background()

	// Create a properly provisioned company.
	created, err := svc.CreateCompany(ctx, company.CreateCompanyRequest{
		Name:        "Invite Test Co",
		InviteEmail: "admin@invite-test.example",
	})
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	userID := uuid.Must(uuid.NewV7())
	testutil.EnsureUser(t, st, userID)

	if err := st.Invite().CreateInvite(ctx, store.CompanyInvite{
		ID: uuid.Must(uuid.NewV7()), CompanyID: created.Company.ID,
		Email: "newmember@example.com", Role: store.InviteRoleSuperAdmin,
		InviteCode: "valid-invite-token", ExpiresAt: now.Add(24 * time.Hour), CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	member, err := svc.AcceptInvite(ctx, company.AcceptInviteRequest{
		UserID:     userID,
		InviteCode: "valid-invite-token",
		Name:       "New Admin",
	})
	if err != nil {
		t.Fatal(err)
	}
	if member.UserID != userID {
		t.Fatalf("expected member.UserID=%s, got %s", userID, member.UserID)
	}
	if member.CompanyID != created.Company.ID {
		t.Fatal("expected member in created company")
	}
	if member.Name != "New Admin" {
		t.Fatalf("expected name 'New Admin', got %s", member.Name)
	}
}

func TestAcceptInviteIdempotent(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	client := &mock.StubAdminClient{User: newapi.User{ID: 701, Quota: 0}}
	svc := company.NewService(cfg, st, newapi.NewAdminPortAdapter(client), permission.NewGrantNormalizer())
	ctx := context.Background()

	// Create a provisioned company.
	created, err := svc.CreateCompany(ctx, company.CreateCompanyRequest{
		Name:        "Idem Co",
		InviteEmail: "admin@idem.example",
	})
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	userID := uuid.Must(uuid.NewV7())
	testutil.EnsureUser(t, st, userID)

	// Create two invites for same company.
	if err := st.Invite().CreateInvite(ctx, store.CompanyInvite{
		ID: uuid.Must(uuid.NewV7()), CompanyID: created.Company.ID,
		Email: "idem@example.com", Role: store.InviteRoleSuperAdmin,
		InviteCode: "idem-1", ExpiresAt: now.Add(24 * time.Hour), CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	if err := st.Invite().CreateInvite(ctx, store.CompanyInvite{
		ID: uuid.Must(uuid.NewV7()), CompanyID: created.Company.ID,
		Email: "idem@example.com", Role: store.InviteRoleSuperAdmin,
		InviteCode: "idem-2", ExpiresAt: now.Add(24 * time.Hour), CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	member1, err := svc.AcceptInvite(ctx, company.AcceptInviteRequest{
		UserID: userID, InviteCode: "idem-1", Name: "Admin",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Second accept with same user → should return existing member (addMember is idempotent).
	member2, err := svc.AcceptInvite(ctx, company.AcceptInviteRequest{
		UserID: userID, InviteCode: "idem-2", Name: "Admin Again",
	})
	if err != nil {
		t.Fatal(err)
	}
	if member1.ID != member2.ID {
		t.Fatalf("expected same member ID (idempotent), got %s vs %s", member1.ID, member2.ID)
	}
}

func TestAcceptInviteRejectsExpiredToken(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	client := &mock.StubAdminClient{User: newapi.User{ID: 703, Quota: 0}}
	svc := company.NewService(cfg, st, newapi.NewAdminPortAdapter(client), permission.NewGrantNormalizer())
	ctx := context.Background()

	created, err := svc.CreateCompany(ctx, company.CreateCompanyRequest{
		Name:        "Expired Co",
		InviteEmail: "admin@expired.example",
	})
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	userID := uuid.Must(uuid.NewV7())
	testutil.EnsureUser(t, st, userID)

	if err := st.Invite().CreateInvite(ctx, store.CompanyInvite{
		ID: uuid.Must(uuid.NewV7()), CompanyID: created.Company.ID,
		Email: "expired@example.com", Role: store.InviteRoleSuperAdmin,
		InviteCode: "expired-token", ExpiresAt: now.Add(-time.Hour), CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	_, err = svc.AcceptInvite(ctx, company.AcceptInviteRequest{
		UserID: userID, InviteCode: "expired-token", Name: "Admin",
	})
	if err == nil {
		t.Fatal("expected error for expired invite")
	}
}

func TestAcceptInviteRejectsAlreadyAccepted(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	client := &mock.StubAdminClient{User: newapi.User{ID: 702, Quota: 0}}
	svc := company.NewService(cfg, st, newapi.NewAdminPortAdapter(client), permission.NewGrantNormalizer())
	ctx := context.Background()

	created, err := svc.CreateCompany(ctx, company.CreateCompanyRequest{
		Name:        "Already Accepted Co",
		InviteEmail: "admin@already.example",
	})
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	userID := uuid.Must(uuid.NewV7())
	testutil.EnsureUser(t, st, userID)

	if err := st.Invite().CreateInvite(ctx, store.CompanyInvite{
		ID: uuid.Must(uuid.NewV7()), CompanyID: created.Company.ID,
		Email: "used@example.com", Role: store.InviteRoleSuperAdmin,
		InviteCode: "used-token", ExpiresAt: now.Add(24 * time.Hour), CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	_, err = svc.AcceptInvite(ctx, company.AcceptInviteRequest{
		UserID: userID, InviteCode: "used-token", Name: "Admin",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create another user to try accepting the same invite.
	userID2 := uuid.Must(uuid.NewV7())
	testutil.EnsureUser(t, st, userID2)
	_, err = svc.AcceptInvite(ctx, company.AcceptInviteRequest{
		UserID: userID2, InviteCode: "used-token", Name: "Other",
	})
	if err == nil {
		t.Fatal("expected error for already accepted invite")
	}
}

func TestAcceptInviteRejectsNilUserID(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := company.NewService(cfg, st, nil, permission.NewGrantNormalizer())

	_, err := svc.AcceptInvite(context.Background(), company.AcceptInviteRequest{
		InviteCode: "any", Name: "Admin",
	})
	if err == nil {
		t.Fatal("expected error for nil userID")
	}
}

func TestAcceptInviteRejectsInvalidToken(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := company.NewService(cfg, st, nil, permission.NewGrantNormalizer())

	userID := uuid.Must(uuid.NewV7())
	testutil.EnsureUser(t, st, userID)

	_, err := svc.AcceptInvite(context.Background(), company.AcceptInviteRequest{
		UserID: userID, InviteCode: "does-not-exist", Name: "Admin",
	})
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}
