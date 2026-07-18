//go:build testhook

package company_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestPendingInvitesForUserByEmail(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := company.NewService(cfg, st, nil, permission.NewGrantNormalizer())
	ctx := context.Background()

	now := time.Now().UTC()
	if err := st.Invite().CreateInvite(ctx, store.CompanyInvite{
		ID: uuid.Must(uuid.NewV7()), CompanyID: contract.DefaultCompanyID,
		Email: "pending@example.com", Role: store.InviteRoleSuperAdmin,
		InviteCode: "pending-email-code", ExpiresAt: now.Add(24 * time.Hour), CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	invites, err := svc.PendingInvitesForUser(ctx, company.PendingInvitesForUserRequest{
		Email: "pending@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(invites) == 0 {
		t.Fatal("expected at least one pending invite")
	}
	found := false
	for _, inv := range invites {
		if inv.InviteCode == "pending-email-code" {
			found = true
			if inv.CompanyID != contract.DefaultCompanyID {
				t.Fatalf("expected companyID=%s, got %s", contract.DefaultCompanyID, inv.CompanyID)
			}
			if inv.CompanyName == "" {
				t.Fatal("expected company name to be populated")
			}
		}
	}
	if !found {
		t.Fatal("expected to find our pending invite")
	}
}

func TestPendingInvitesForUserByPhone(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := company.NewService(cfg, st, nil, permission.NewGrantNormalizer())
	ctx := context.Background()

	now := time.Now().UTC()
	if err := st.Invite().CreateInvite(ctx, store.CompanyInvite{
		ID: uuid.Must(uuid.NewV7()), CompanyID: contract.DefaultCompanyID,
		Phone: "13800138000", Role: "member",
		InviteCode: "pending-phone-code", ExpiresAt: now.Add(24 * time.Hour), CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	invites, err := svc.PendingInvitesForUser(ctx, company.PendingInvitesForUserRequest{
		Phone: "13800138000",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(invites) == 0 {
		t.Fatal("expected pending invite by phone")
	}
	if invites[0].InviteCode != "pending-phone-code" {
		t.Fatalf("expected pending-phone-code, got %s", invites[0].InviteCode)
	}
}

func TestPendingInvitesForUserExcludesExpired(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := company.NewService(cfg, st, nil, permission.NewGrantNormalizer())
	ctx := context.Background()

	now := time.Now().UTC()
	// Expired invite.
	if err := st.Invite().CreateInvite(ctx, store.CompanyInvite{
		ID: uuid.Must(uuid.NewV7()), CompanyID: contract.DefaultCompanyID,
		Email: "expired-pending@example.com", Role: store.InviteRoleSuperAdmin,
		InviteCode: "expired-pending", ExpiresAt: now.Add(-time.Hour), CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	invites, err := svc.PendingInvitesForUser(ctx, company.PendingInvitesForUserRequest{
		Email: "expired-pending@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, inv := range invites {
		if inv.InviteCode == "expired-pending" {
			t.Fatal("expected expired invite to be excluded")
		}
	}
}

func TestPendingInvitesForUserExcludesAccepted(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := company.NewService(cfg, st, nil, permission.NewGrantNormalizer())
	ctx := context.Background()

	now := time.Now().UTC()
	inviteID := uuid.Must(uuid.NewV7())
	if err := st.Invite().CreateInvite(ctx, store.CompanyInvite{
		ID: inviteID, CompanyID: contract.DefaultCompanyID,
		Email: "accepted-pending@example.com", Role: store.InviteRoleSuperAdmin,
		InviteCode: "accepted-pending", ExpiresAt: now.Add(24 * time.Hour), CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	if err := st.Invite().MarkInviteAccepted(ctx, inviteID, now); err != nil {
		t.Fatal(err)
	}

	invites, err := svc.PendingInvitesForUser(ctx, company.PendingInvitesForUserRequest{
		Email: "accepted-pending@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, inv := range invites {
		if inv.InviteCode == "accepted-pending" {
			t.Fatal("expected accepted invite to be excluded")
		}
	}
}

func TestPendingInvitesForUserEmptyReturnsNil(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	svc := company.NewService(cfg, st, nil, permission.NewGrantNormalizer())

	invites, err := svc.PendingInvitesForUser(context.Background(), company.PendingInvitesForUserRequest{
		Email: "nobody@nowhere.example",
	})
	if err != nil {
		t.Fatal(err)
	}
	if invites != nil {
		t.Fatalf("expected nil for no invites, got %v", invites)
	}
}
