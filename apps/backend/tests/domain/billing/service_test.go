package billing_test

import (
	"context"
	"testing"

	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	"github.com/tokenjoy/backend/internal/domain/company"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func newBillingService(t *testing.T) (domainbilling.Service, store.Store, context.Context) {
	t.Helper()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	reader := domainusage.NewReader(st.Usage(), st.Ledger())
	svc := domainbilling.NewService(cfg, st, reader)
	co, err := st.Company().GetByID(context.Background(), contract.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	walletID := int64(501)
	if err := st.Company().UpdateNewAPIWalletCompanyID(context.Background(), contract.DefaultCompanyID, walletID); err != nil {
		t.Fatal(err)
	}
	ctx := company.WithContext(context.Background(), company.Context{
		CompanyID:             contract.DefaultCompanyID,
		NewAPIWalletCompanyID: walletID,
		Status:                co.Status,
	})
	return svc, st, ctx
}

func TestGetWalletReturnsBalance(t *testing.T) {
	t.Parallel()
	svc, _, ctx := newBillingService(t)
	if err := svc.PlatformRecharge(ctx, contract.DefaultCompanyID, 100, contract.IDMemberAdmin); err != nil {
		t.Fatal(err)
	}
	view, err := svc.GetWallet(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if domainbilling.PrimaryWalletBalance(view) <= 0 {
		t.Fatalf("expected positive balance, got %v", domainbilling.PrimaryWalletBalance(view))
	}
	if view.CompanyID != contract.DefaultCompanyID {
		t.Fatalf("expected company %d, got %d", contract.DefaultCompanyID, view.CompanyID)
	}
}

func TestGetWalletWithoutNewAPIWalletCompanyIDReturnsZero(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	reader := domainusage.NewReader(st.Usage(), st.Ledger())
	svc := domainbilling.NewService(cfg, st, reader)
	ctx := testutil.Ctx()
	view, err := svc.GetWallet(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if domainbilling.PrimaryWalletBalance(view) != 0 {
		t.Fatalf("expected zero balance, got %v", domainbilling.PrimaryWalletBalance(view))
	}
}

func TestConfirmPaymentIdempotent(t *testing.T) {
	t.Parallel()
	svc, _, ctx := newBillingService(t)
	order, err := svc.CreateSelfRecharge(ctx, 20, "idem-key-1", contract.IDMemberAdmin)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.ConfirmPayment(ctx, order.ID); err != nil {
		t.Fatal(err)
	}
	if err := svc.ConfirmPayment(ctx, order.ID); err != nil {
		t.Fatal(err)
	}
}

func TestCreateSelfRechargeRejectsDuplicateIdempotencyKey(t *testing.T) {
	t.Parallel()
	svc, _, ctx := newBillingService(t)
	if _, err := svc.CreateSelfRecharge(ctx, 10, "dup-key", contract.IDMemberAdmin); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateSelfRecharge(ctx, 10, "dup-key", contract.IDMemberAdmin); err == nil {
		t.Fatal("expected error for duplicate idempotency key")
	}
}

func TestCreateSelfRechargeUsesCurrenciesPointsPerUnit(t *testing.T) {
	t.Parallel()
	svc, st, ctx := newBillingService(t)
	cur, err := st.Billing().GetCurrency(ctx, common.DefaultBillingCurrency)
	if err != nil {
		t.Fatal(err)
	}
	if cur == nil || !cur.Enabled || cur.QuotaPerUnit <= 0 {
		t.Fatalf("expected seeded CNY currency, got %+v", cur)
	}
	order, err := svc.CreateSelfRecharge(ctx, 15, "ppu-key-1", contract.IDMemberAdmin)
	if err != nil {
		t.Fatal(err)
	}
	if order.Currency != common.DefaultBillingCurrency {
		t.Fatalf("currency: got %q want default currency", order.Currency)
	}
	if order.QuotaPerUnit != cur.QuotaPerUnit {
		t.Fatalf("points_per_unit: got %d want %d (from currencies)", order.QuotaPerUnit, cur.QuotaPerUnit)
	}
	wantGranted := common.QuotaFromAmount(15, cur.QuotaPerUnit)
	if order.QuotaGranted != wantGranted {
		t.Fatalf("points_granted: got %v want %v", order.QuotaGranted, wantGranted)
	}
}
