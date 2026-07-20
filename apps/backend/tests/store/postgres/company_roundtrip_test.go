package postgres_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	billinglot "github.com/tokenjoy/backend/internal/domain/billing/lot"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestCompanyRoundTrip(t *testing.T) {
	t.Parallel()
	st := testPostgresStore(t)
	ctx := testutil.Ctx()
	now := time.Now().UTC()

	co := store.Company{
		ID: uuid.MustParse("00000000-0000-7000-0000-000000009001"), Name: "Roundtrip Co",
		Type: store.CompanyTypeTesting, Status: store.CompanyStatusActive,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := st.Company().Create(ctx, co); err != nil {
		t.Fatal(err)
	}
	got, err := st.Company().GetByID(ctx, co.ID)
	if err != nil || got == nil {
		t.Fatal("expected company")
	}
	if got.Name != co.Name {
		t.Fatalf("expected name %s, got %s", co.Name, got.Name)
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
		ID: uuid.MustParse("00000000-0000-7000-0000-0000000000a1"), CompanyID: contract.DefaultCompanyID, Email: "rt@example.com",
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
	ppu := domainbilling.DefaultQuotaPerUnit()

	order := store.RechargeOrder{
		ID: uuid.MustParse("00000000-0000-7000-0000-0000000000b1"), CompanyID: contract.DefaultCompanyID, Amount: 99, Currency: common.DefaultBillingCurrency,
		QuotaPerUnit: ppu, QuotaGranted: common.QuotaFromAmount(99, ppu),
		Source: store.RechargeSourceSelf, LotKind: store.LotKindPaid,
		IdempotencyKey: &key, Status: store.RechargeStatusPending,
		DisplayOrderID: "ORD20260101120000",
		PaymentMethod:  store.PaymentMethodAlipay,
		InvoiceStatus:  store.InvoiceStatusNone,
		CreatedBy:      contract.IDMemberAdmin, CreatedAt: now, UpdatedAt: now,
	}
	if err := st.Billing().CreateRechargeOrder(ctx, order); err != nil {
		t.Fatal(err)
	}
	got, err := st.Billing().GetRechargeOrder(ctx, order.ID)
	if err != nil || got == nil || got.Amount != 99 {
		t.Fatalf("unexpected order: %+v err=%v", got, err)
	}
	order.Status = store.RechargeStatusConfirmed
	lot := domainbilling.BuildPaidLot(order, common.DefaultBillingCurrency)
	before, err := st.Company().GetByID(ctx, order.CompanyID)
	if err != nil || before == nil {
		t.Fatal("expected company before recharge")
	}
	if err := billinglot.CreditFromLot(ctx, st, order, lot, lot.QuotaGranted); err != nil {
		t.Fatal(err)
	}
	got, err = st.Billing().GetRechargeOrder(ctx, order.ID)
	if err != nil || got.Status != store.RechargeStatusConfirmed {
		t.Fatalf("expected confirmed, got %+v", got)
	}
	orders, err := st.Billing().ListRechargeOrders(ctx, contract.DefaultCompanyID)
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
	if co.WalletRemain != before.WalletRemain+lot.QuotaGranted {
		t.Fatalf("wallet_remain: got %v want %v", co.WalletRemain, before.WalletRemain+lot.QuotaGranted)
	}
}
