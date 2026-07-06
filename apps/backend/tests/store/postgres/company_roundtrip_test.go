//go:build integration

package postgres_test

import (
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestCompanyRoundTrip(t *testing.T) {
	st := testPostgresStore(t)
	ctx := testutil.Ctx()
	now := time.Now().UTC()

	co := store.Company{
		ID: 9001, Slug: "roundtrip-co", Name: "Roundtrip Co",
		Status: store.CompanyStatusActive, CreatedAt: now, UpdatedAt: now,
	}
	if err := st.Company().Create(ctx, co); err != nil {
		t.Fatal(err)
	}
	got, err := st.Company().GetByID(ctx, co.ID)
	if err != nil || got == nil {
		t.Fatal("expected company")
	}
	if got.Slug != co.Slug {
		t.Fatalf("expected slug %s, got %s", co.Slug, got.Slug)
	}
	if err := st.Company().UpdateStatus(ctx, co.ID, store.CompanyStatusSuspended); err != nil {
		t.Fatal(err)
	}
	got, err = st.Company().GetByID(ctx, co.ID)
	if err != nil || got.Status != store.CompanyStatusSuspended {
		t.Fatalf("expected suspended, got %+v", got)
	}
}

func TestCompanyInviteRoundTrip(t *testing.T) {
	st := testPostgresStore(t)
	ctx := testutil.Ctx()
	now := time.Now().UTC()

	invite := store.CompanyInvite{
		ID: "invite-rt-1", CompanyID: 1, Email: "rt@example.com",
		Role: store.InviteRoleSuperAdmin, Token: "roundtrip-invite-token",
		ExpiresAt: now.Add(24 * time.Hour), CreatedAt: now,
	}
	if err := st.Invite().CreateInvite(ctx, invite); err != nil {
		t.Fatal(err)
	}
	got, err := st.Invite().GetInviteByToken(ctx, invite.Token)
	if err != nil || got == nil || got.Email != invite.Email {
		t.Fatalf("unexpected invite: %+v err=%v", got, err)
	}
	accepted := now.Add(time.Minute)
	if err := st.Invite().MarkInviteAccepted(ctx, invite.ID, accepted); err != nil {
		t.Fatal(err)
	}
	got, err = st.Invite().GetInviteByToken(ctx, invite.Token)
	if err != nil || got.AcceptedAt == nil {
		t.Fatal("expected accepted_at to be set")
	}
}

func TestRechargeOrderRoundTrip(t *testing.T) {
	st := testPostgresStore(t)
	ctx := testutil.Ctx()
	now := time.Now().UTC()
	key := "idem-rt-key"

	order := store.RechargeOrder{
		ID: "rch-rt-1", CompanyID: 1, Amount: 99, Source: store.RechargeSourceSelf,
		IdempotencyKey: &key, Status: store.RechargeStatusPending,
		DisplayOrderID: "ORD20260101120000",
		PaymentMethod:  store.PaymentMethodAlipay,
		InvoiceStatus:  store.InvoiceStatusNone,
		CreatedBy: "m-admin", CreatedAt: now, UpdatedAt: now,
	}
	if err := st.Billing().CreateRechargeOrder(ctx, order); err != nil {
		t.Fatal(err)
	}
	got, err := st.Billing().GetRechargeOrder(ctx, order.ID)
	if err != nil || got == nil || got.Amount != 99 {
		t.Fatalf("unexpected order: %+v err=%v", got, err)
	}
	ref := "topup-ref"
	if err := st.Billing().UpdateRechargeStatus(ctx, order.ID, store.RechargeStatusToppedUp, &ref); err != nil {
		t.Fatal(err)
	}
	got, err = st.Billing().GetRechargeOrder(ctx, order.ID)
	if err != nil || got.Status != store.RechargeStatusToppedUp {
		t.Fatalf("expected topped_up, got %+v", got)
	}
	orders, err := st.Billing().ListRechargeOrders(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, o := range orders {
		if o.ID == order.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected order in company list")
	}
}
