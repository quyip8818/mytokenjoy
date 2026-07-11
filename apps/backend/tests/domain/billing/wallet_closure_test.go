package billing_test

import (
	"context"
	"fmt"
	"math"
	"testing"

	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	"github.com/tokenjoy/backend/internal/domain/company"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
	riverfix "github.com/tokenjoy/backend/tests/testutil/river"
)

func newBillingServiceWithSync(t *testing.T, client *mock.StubAdminClient) (domainbilling.Service, store.Store, context.Context) {
	t.Helper()
	cfg, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))
	enqueuer := riverfix.NewInsertOnlyEnqueuer(t, cfg, st)
	wallet := company.NewWalletService(cfg, client)
	reader := domainusage.NewReader(st.Usage(), st.Ledger())
	svc := domainbilling.NewService(cfg, st, reader, newapi.NewAdminPortAdapter(client), wallet,
		func(ctx context.Context, companyID int64) error {
			return jobs.InsertRebalance(ctx, enqueuer, nil, companyID, store.RebalanceAxisCompany, fmt.Sprintf("%d", companyID))
		},
		func(ctx context.Context, companyID int64) error {
			return jobs.InsertWalletSync(ctx, enqueuer, nil, companyID)
		},
	)
	co, err := st.Company().GetByID(context.Background(), contract.DefaultCompanyID)
	if err != nil {
		t.Fatal(err)
	}
	walletID := int64(701)
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

func TestWalletClosureFormula(t *testing.T) {
	t.Parallel()
	client := &mock.StubAdminClient{
		GetUserQuotaFn: func(_ context.Context, _ int64) (int64, error) { return 0, nil },
	}
	svc, _, ctx := newBillingServiceWithSync(t, client)

	if err := svc.PlatformRecharge(ctx, contract.DefaultCompanyID, 100, "platform-op-closure"); err != nil {
		t.Fatal(err)
	}
	if err := svc.PlatformGift(ctx, contract.DefaultCompanyID, 5000, "platform-op-gift"); err != nil {
		t.Fatal(err)
	}
	if err := svc.PlatformAdjust(ctx, contract.DefaultCompanyID, 2000, 2, "platform-op-adjust"); err != nil {
		t.Fatal(err)
	}

	view, err := svc.GetWallet(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(view.Balances) == 0 {
		t.Fatal("expected balances")
	}
	for _, b := range view.Balances {
		delta := b.TotalTopup - b.TotalConsumed - b.Balance
		if math.Abs(delta) > 1e-6 {
			t.Fatalf("closure failed for %s: topup(%v) - consumed(%v) != balance(%v), delta=%v",
				b.Currency, b.TotalTopup, b.TotalConsumed, b.Balance, delta)
		}
	}
	if view.BalancePoint <= 0 {
		t.Fatalf("expected positive balancePoint, got %v", view.BalancePoint)
	}
	if view.GiftPoints != 5000 {
		t.Fatalf("expected giftPoints=5000, got %v", view.GiftPoints)
	}
}
