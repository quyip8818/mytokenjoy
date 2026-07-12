package postgres_test

import (
	"testing"
	"time"

	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	domainwallet "github.com/tokenjoy/backend/internal/domain/wallet"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestCompanyRoundTrip(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	st := testPostgresStore(t)
	ctx := testutil.Ctx()
	now := time.Now().UTC()

	invite := store.CompanyInvite{
		ID: "invite-rt-1", CompanyID: 1, Email: "rt@example.com",
		Role: store.InviteRoleSuperAdmin, InviteCode: "roundtrip-invite-token",
		ExpiresAt: now.Add(24 * time.Hour), CreatedAt: now,
	}
	if err := st.Invite().CreateInvite(ctx, invite); err != nil {
		t.Fatal(err)
	}
	got, err := st.Invite().GetInviteByCode(ctx, invite.InviteCode)
	if err != nil || got == nil || got.Email != invite.Email {
		t.Fatalf("unexpected invite: %+v err=%v", got, err)
	}
	accepted := now.Add(time.Minute)
	if err := st.Invite().MarkInviteAccepted(ctx, invite.ID, accepted); err != nil {
		t.Fatal(err)
	}
	got, err = st.Invite().GetInviteByCode(ctx, invite.InviteCode)
	if err != nil || got.AcceptedAt == nil {
		t.Fatal("expected accepted_at to be set")
	}
}

func TestRechargeOrderRoundTrip(t *testing.T) {
	t.Parallel()
	st := testPostgresStore(t)
	ctx := testutil.Ctx()
	now := time.Now().UTC()
	key := "idem-rt-key"
	ppu := domainbilling.DefaultPointsPerUnit()

	order := store.RechargeOrder{
		ID: "rch-rt-1", CompanyID: 1, Amount: 99, Currency: "CNY",
		PointsPerUnit: ppu, PointsGranted: domainbilling.PointsGrantedFromAmount(99, ppu),
		Source: store.RechargeSourceSelf, LotKind: store.LotKindPaid,
		IdempotencyKey: &key, Status: store.RechargeStatusPending,
		DisplayOrderID: "ORD20260101120000",
		PaymentMethod:  store.PaymentMethodAlipay,
		InvoiceStatus:  store.InvoiceStatusNone,
		CreatedBy:      "m-admin", CreatedAt: now, UpdatedAt: now,
	}
	if err := st.Billing().CreateRechargeOrder(ctx, order); err != nil {
		t.Fatal(err)
	}
	got, err := st.Billing().GetRechargeOrder(ctx, order.ID)
	if err != nil || got == nil || got.Amount != 99 {
		t.Fatalf("unexpected order: %+v err=%v", got, err)
	}
	order.Status = store.RechargeStatusConfirmed
	lot := domainbilling.BuildPaidLot(order, "CNY", ppu)
	before, err := st.Company().GetByID(ctx, order.CompanyID)
	if err != nil || before == nil {
		t.Fatal("expected company before recharge")
	}
	if err := domainwallet.CreditFromLot(ctx, st, order, lot, lot.PointsGranted); err != nil {
		t.Fatal(err)
	}
	got, err = st.Billing().GetRechargeOrder(ctx, order.ID)
	if err != nil || got.Status != store.RechargeStatusConfirmed {
		t.Fatalf("expected confirmed, got %+v", got)
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
	co, err := st.Company().GetByID(ctx, order.CompanyID)
	if err != nil || co == nil {
		t.Fatal("expected company after recharge")
	}
	if co.WalletRemain != before.WalletRemain+lot.PointsGranted {
		t.Fatalf("wallet_remain: got %v want %v", co.WalletRemain, before.WalletRemain+lot.PointsGranted)
	}
}
