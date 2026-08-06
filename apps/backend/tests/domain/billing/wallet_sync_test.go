package billing_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	billinglot "github.com/tokenjoy/backend/internal/domain/billing/lot"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/support/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

// recordingSyncer records ManageUser calls for assertion.
type recordingSyncer struct {
	mu    sync.Mutex
	calls []manageCall
}

type manageCall struct {
	UserID int64
	Action string
	Value  int64
}

func (r *recordingSyncer) ManageUser(_ context.Context, userID int64, action string, value int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, manageCall{UserID: userID, Action: action, Value: value})
	return nil
}

func (r *recordingSyncer) lastCall() (manageCall, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.calls) == 0 {
		return manageCall{}, false
	}
	return r.calls[len(r.calls)-1], true
}

func TestCreditFromLotReturnsNewWalletRemain(t *testing.T) {
	t.Parallel()
	companyID := uuid.MustParse("00000000-0000-7000-0000-000000009301")
	_, st := testutil.NewTestStore(t)
	ctx := newLotTestCompany(t, st, companyID)
	now := time.Now().UTC()

	order := paidRechargeOrder(companyID, uuid.MustParse("00000000-0000-7000-0000-000000001101"), 100, now)
	lot := domainbilling.BuildLot(order, common.DefaultBillingCurrency, store.LotKindPaid, order.Amount)

	newRemain, err := billinglot.CreditFromLot(ctx, st, order, lot, lot.QuotaGranted)
	if err != nil {
		t.Fatal(err)
	}
	if newRemain != lot.QuotaGranted {
		t.Fatalf("newRemain: got %d, want %d", newRemain, lot.QuotaGranted)
	}

	// Second recharge should accumulate.
	order2 := paidRechargeOrder(companyID, uuid.MustParse("00000000-0000-7000-0000-000000001102"), 50, now.Add(time.Second))
	lot2 := domainbilling.BuildLot(order2, common.DefaultBillingCurrency, store.LotKindPaid, order2.Amount)

	newRemain2, err := billinglot.CreditFromLot(ctx, st, order2, lot2, lot2.QuotaGranted)
	if err != nil {
		t.Fatal(err)
	}
	want := lot.QuotaGranted + lot2.QuotaGranted
	if newRemain2 != want {
		t.Fatalf("newRemain after 2nd credit: got %d, want %d", newRemain2, want)
	}
}

func TestConsumeLotsDecrementsWallet(t *testing.T) {
	t.Parallel()
	companyID := uuid.MustParse("00000000-0000-7000-0000-000000009302")
	_, st := testutil.NewTestStore(t)
	ctx := newLotTestCompany(t, st, companyID)
	now := time.Now().UTC()

	// Seed balance.
	order := paidRechargeOrder(companyID, uuid.MustParse("00000000-0000-7000-0000-000000001201"), 100, now)
	lot := domainbilling.BuildLot(order, common.DefaultBillingCurrency, store.LotKindPaid, order.Amount)
	_, err := billinglot.CreditFromLot(ctx, st, order, lot, lot.QuotaGranted)
	if err != nil {
		t.Fatal(err)
	}

	// Consume partial.
	consume := lot.QuotaGranted / 4
	_, err = billinglot.ConsumeLots(ctx, st, companyID, consume)
	if err != nil {
		t.Fatal(err)
	}
	co, err := st.Company().GetByID(ctx, companyID)
	if err != nil || co == nil {
		t.Fatal("expected company after consume")
	}
	want := lot.QuotaGranted - consume
	if co.WalletRemainQuota != want {
		t.Fatalf("WalletRemainQuota: got %d, want %d", co.WalletRemainQuota, want)
	}
}

func TestBillingServiceSyncsWalletAfterRecharge(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t, testutil.WithNewAPIEnabled(true))

	// Set up the wallet user ID on the default company.
	walletUserID := int64(999)
	if err := st.Company().UpdateNewAPIWalletCompanyID(context.Background(), contract.DefaultCompanyID, walletUserID); err != nil {
		t.Fatal(err)
	}

	syncer := &recordingSyncer{}
	reader := usage.NewReader(st.Usage(), st.Ledger())
	svc := domainbilling.NewService(st, reader, syncer)

	ctx := company.WithContext(context.Background(), company.Context{CompanyID: contract.DefaultCompanyID})
	if err := svc.PlatformRecharge(ctx, contract.DefaultCompanyID, 200, contract.IDMemberAdmin); err != nil {
		t.Fatal(err)
	}

	call, ok := syncer.lastCall()
	if !ok {
		t.Fatal("expected ManageUser call after recharge")
	}
	if call.UserID != walletUserID {
		t.Fatalf("userID: got %d, want %d", call.UserID, walletUserID)
	}
	if call.Action != "set_quota" {
		t.Fatalf("action: got %q, want %q", call.Action, "set_quota")
	}
	if call.Value <= 0 {
		t.Fatalf("expected positive quota value, got %d", call.Value)
	}
}
