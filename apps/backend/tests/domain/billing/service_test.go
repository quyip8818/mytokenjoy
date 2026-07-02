package billing_test

import (
	"context"
	"fmt"
	"testing"

	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/memory"
	"github.com/tokenjoy/backend/internal/store/seed"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func newBillingService(t *testing.T, client *mock.StubAdminClient) (domainbilling.Service, store.Store, context.Context) {
	t.Helper()
	cfg := testutil.TestConfig(testutil.WithNewAPIEnabled(true))
	st := memory.New(seed.Load(cfg))
	wallet := company.NewWalletService(cfg, client)
	svc := domainbilling.NewService(cfg, st, client, wallet, func(ctx context.Context, companyID int64) error {
		return st.Relay().EnqueueRebalance(ctx, store.RebalanceAxisCompany, fmt.Sprintf("%d", companyID))
	})
	co, err := st.Company().GetByID(context.Background(), seed.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	walletID := int64(501)
	if err := st.Company().UpdateNewAPIWalletUserID(context.Background(), seed.DefaultCompanyID, walletID); err != nil {
		t.Fatal(err)
	}
	ctx := company.WithContext(context.Background(), company.Context{
		CompanyID:          seed.DefaultCompanyID,
		NewAPIWalletUserID: walletID,
		Status:             co.Status,
	})
	return svc, st, ctx
}

func TestGetWalletReturnsBalance(t *testing.T) {
	client := &mock.StubAdminClient{
		GetUserQuotaFn: func(_ context.Context, userID int64) (int64, error) {
			return 1_000_000, nil
		},
	}
	svc, _, ctx := newBillingService(t, client)
	view, err := svc.GetWallet(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if view.Balance <= 0 {
		t.Fatalf("expected positive balance, got %v", view.Balance)
	}
	if view.CompanyID != seed.DefaultCompanyID {
		t.Fatalf("expected company %d, got %d", seed.DefaultCompanyID, view.CompanyID)
	}
}

func TestGetWalletWithoutNewAPIWalletUserIDReturnsZero(t *testing.T) {
	cfg := testutil.TestConfig(testutil.WithNewAPIEnabled(true))
	st := memory.New(seed.Load(cfg))
	client := &mock.StubAdminClient{}
	svc := domainbilling.NewService(cfg, st, client, company.NewWalletService(cfg, client), nil)
	ctx := testutil.Ctx()
	view, err := svc.GetWallet(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if view.Balance != 0 {
		t.Fatalf("expected zero balance, got %v", view.Balance)
	}
}

func TestPlatformRechargeEnqueuesRebalance(t *testing.T) {
	client := &mock.StubAdminClient{
		GetUserQuotaFn: func(_ context.Context, _ int64) (int64, error) { return 0, nil },
	}
	svc, st, ctx := newBillingService(t, client)
	if err := svc.PlatformRecharge(ctx, seed.DefaultCompanyID, 50, "platform-op-1"); err != nil {
		t.Fatal(err)
	}
	if testutil.PendingRebalanceCount(st, seed.DefaultCompanyID) == 0 {
		t.Fatal("expected rebalance outbox entry after platform recharge")
	}
}

func TestConfirmPaymentIdempotent(t *testing.T) {
	client := &mock.StubAdminClient{
		GetUserQuotaFn: func(_ context.Context, _ int64) (int64, error) { return 0, nil },
	}
	svc, _, ctx := newBillingService(t, client)
	order, err := svc.CreateSelfRecharge(ctx, 20, "idem-key-1", "m-admin")
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
	client := &mock.StubAdminClient{}
	svc, _, ctx := newBillingService(t, client)
	if _, err := svc.CreateSelfRecharge(ctx, 10, "dup-key", "m-admin"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateSelfRecharge(ctx, 10, "dup-key", "m-admin"); err == nil {
		t.Fatal("expected error for duplicate idempotency key")
	}
}

func TestConfirmPaymentFailsWithoutWallet(t *testing.T) {
	cfg := testutil.TestConfig(testutil.WithNewAPIEnabled(true))
	st := memory.New(seed.Load(cfg))
	client := &mock.StubAdminClient{}
	svc := domainbilling.NewService(cfg, st, client, company.NewWalletService(cfg, client), nil)
	ctx := testutil.Ctx()
	order, err := svc.CreateSelfRecharge(ctx, 15, "no-wallet-key", "m-admin")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.ConfirmPayment(ctx, order.ID); err == nil {
		t.Fatal("expected error when wallet not configured")
	}
	stored, err := st.Billing().GetRechargeOrder(ctx, order.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != store.RechargeStatusFailed {
		t.Fatalf("expected failed status, got %s", stored.Status)
	}
}
