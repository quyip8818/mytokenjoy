package billing_test

import (
	"context"
	"fmt"
	"testing"

	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	"github.com/tokenjoy/backend/internal/domain/company"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func newBillingService(t *testing.T, client *mock.StubAdminClient) (domainbilling.Service, store.Store, context.Context) {
	t.Helper()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	wallet := company.NewWalletService(cfg, client)
	reader := domainusage.NewReader(st.Usage(), st.Ledger())
	svc := domainbilling.NewService(cfg, st, reader, client, wallet,
		func(ctx context.Context, companyID int64) error {
			return st.Relay().EnqueueRebalance(ctx, store.RebalanceAxisCompany, fmt.Sprintf("%d", companyID))
		},
		nil,
	)
	co, err := st.Company().GetByID(context.Background(), contract.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	walletID := int64(501)
	if err := st.Company().UpdateNewAPIWalletUserID(context.Background(), contract.DefaultCompanyID, walletID); err != nil {
		t.Fatal(err)
	}
	ctx := company.WithContext(context.Background(), company.Context{
		CompanyID:          contract.DefaultCompanyID,
		NewAPIWalletUserID: walletID,
		Status:             co.Status,
	})
	return svc, st, ctx
}

func TestGetWalletReturnsBalance(t *testing.T) {
	t.Parallel()
	client := &mock.StubAdminClient{
		GetUserQuotaFn: func(_ context.Context, userID int64) (int64, error) {
			return 1_000_000, nil
		},
	}
	svc, _, ctx := newBillingService(t, client)
	if err := svc.PlatformRecharge(ctx, contract.DefaultCompanyID, 100, "platform-op-wallet"); err != nil {
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

func TestGetWalletWithoutNewAPIWalletUserIDReturnsZero(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	client := &mock.StubAdminClient{}
	reader := domainusage.NewReader(st.Usage(), st.Ledger())
	svc := domainbilling.NewService(cfg, st, reader, client, company.NewWalletService(cfg, client), nil, nil)
	ctx := testutil.Ctx()
	view, err := svc.GetWallet(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if domainbilling.PrimaryWalletBalance(view) != 0 {
		t.Fatalf("expected zero balance, got %v", domainbilling.PrimaryWalletBalance(view))
	}
}

func TestPlatformRechargeEnqueuesRebalance(t *testing.T) {
	t.Parallel()
	client := &mock.StubAdminClient{
		GetUserQuotaFn: func(_ context.Context, _ int64) (int64, error) { return 0, nil },
	}
	svc, st, ctx := newBillingService(t, client)
	if err := svc.PlatformRecharge(ctx, contract.DefaultCompanyID, 50, "platform-op-1"); err != nil {
		t.Fatal(err)
	}
	if testutil.PendingRebalanceCount(st, contract.DefaultCompanyID) == 0 {
		t.Fatal("expected rebalance outbox entry after platform recharge")
	}
}

func TestConfirmPaymentIdempotent(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	client := &mock.StubAdminClient{}
	svc, _, ctx := newBillingService(t, client)
	if _, err := svc.CreateSelfRecharge(ctx, 10, "dup-key", "m-admin"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateSelfRecharge(ctx, 10, "dup-key", "m-admin"); err == nil {
		t.Fatal("expected error for duplicate idempotency key")
	}
}

func TestConfirmPaymentSucceedsWithoutNewAPIWallet(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(false))
	client := &mock.StubAdminClient{}
	reader := domainusage.NewReader(st.Usage(), st.Ledger())
	svc := domainbilling.NewService(cfg, st, reader, client, company.NewWalletService(cfg, client), nil, nil)
	ctx := testutil.Ctx()
	order, err := svc.CreateSelfRecharge(ctx, 15, "no-wallet-key", "m-admin")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.ConfirmPayment(ctx, order.ID); err != nil {
		t.Fatal(err)
	}
	stored, err := st.Billing().GetRechargeOrder(ctx, order.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != store.RechargeStatusConfirmed {
		t.Fatalf("expected confirmed status, got %s", stored.Status)
	}
	wallet, err := svc.GetWallet(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if domainbilling.PrimaryWalletBalance(wallet) <= 0 {
		t.Fatalf("expected positive balance after confirm, got %v", domainbilling.PrimaryWalletBalance(wallet))
	}
}

func TestPlatformRechargeSucceedsWhenNewAPIDisabled(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(false))
	client := &mock.StubAdminClient{}
	reader := domainusage.NewReader(st.Usage(), st.Ledger())
	svc := domainbilling.NewService(cfg, st, reader, client, company.NewWalletService(cfg, client), nil, nil)
	ctx := testutil.Ctx()
	if err := svc.PlatformRecharge(ctx, contract.DefaultCompanyID, 25, "platform-op-disabled"); err != nil {
		t.Fatal(err)
	}
	orders, err := st.Billing().ListRechargeOrders(ctx, contract.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	if len(orders) == 0 {
		t.Fatal("expected recharge order to be created")
	}
	last := orders[0]
	if last.Status != store.RechargeStatusConfirmed {
		t.Fatalf("expected confirmed status, got %s", last.Status)
	}
	if client.TopUpCalls != 0 {
		t.Fatalf("expected no TopUp calls, got %d", client.TopUpCalls)
	}
}
