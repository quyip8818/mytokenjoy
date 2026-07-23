package billing_test

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

func TestSeedTrialCreditCreatesTrialLot(t *testing.T) {
	t.Parallel()
	companyID := uuid.MustParse("00000000-0000-7000-0000-000000009201")
	_, st := testutil.NewTestStore(t)
	ctx := newLotTestCompany(t, st, companyID)

	trialQuota := int64(10000)
	if err := domainbilling.SeedTrialCredit(ctx, st, companyID, trialQuota, nil); err != nil {
		t.Fatal(err)
	}

	// Verify wallet_quota_remain is credited.
	co, err := st.Company().GetByID(ctx, companyID)
	if err != nil || co == nil {
		t.Fatal("expected company after trial credit")
	}
	if co.WalletQuotaRemain != trialQuota {
		t.Fatalf("wallet_quota_remain: got %v want %v", co.WalletQuotaRemain, trialQuota)
	}

	// Verify lot exists with correct kind.
	lots, err := st.Billing().ListActiveLotsFIFO(ctx, companyID, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(lots) != 1 {
		t.Fatalf("expected 1 active lot, got %d", len(lots))
	}
	if lots[0].LotKind != store.LotKindMock {
		t.Fatalf("lot kind: got %q want %q", lots[0].LotKind, store.LotKindMock)
	}
	if lots[0].QuotaGranted != trialQuota {
		t.Fatalf("quota granted: got %v want %v", lots[0].QuotaGranted, trialQuota)
	}
	if lots[0].AmountDisplay != 0 {
		t.Fatalf("amount display should be 0 for trial lot, got %v", lots[0].AmountDisplay)
	}
}

func TestSeedTrialCreditRejectsZeroPoints(t *testing.T) {
	t.Parallel()
	companyID := uuid.MustParse("00000000-0000-7000-0000-000000009202")
	_, st := testutil.NewTestStore(t)
	_ = newLotTestCompany(t, st, companyID)
	ctx := testutil.CtxForCompany(companyID)

	if err := domainbilling.SeedTrialCredit(ctx, st, companyID, 0, nil); err == nil {
		t.Fatal("expected error for zero trial points")
	}
	if err := domainbilling.SeedTrialCredit(ctx, st, companyID, -100, nil); err == nil {
		t.Fatal("expected error for negative trial points")
	}
}

func TestExpireMockLotsZerosWalletQuotaRemain(t *testing.T) {
	t.Parallel()
	companyID := uuid.MustParse("00000000-0000-7000-0000-000000009203")
	_, st := testutil.NewTestStore(t)
	ctx := newLotTestCompany(t, st, companyID)

	// Seed trial credit.
	trialQuota := int64(10000)
	if err := domainbilling.SeedTrialCredit(ctx, st, companyID, trialQuota, nil); err != nil {
		t.Fatal(err)
	}

	// Verify wallet is credited.
	co, err := st.Company().GetByID(ctx, companyID)
	if err != nil || co == nil {
		t.Fatal("expected company")
	}
	if co.WalletQuotaRemain != trialQuota {
		t.Fatalf("before expire: wallet_quota_remain got %v want %v", co.WalletQuotaRemain, trialQuota)
	}

	// Expire trial lots (simulates upgrade).
	if err := domainbilling.ExpireMockLots(ctx, st, companyID); err != nil {
		t.Fatal(err)
	}

	// Wallet should be zero — no real lots remain.
	co, err = st.Company().GetByID(ctx, companyID)
	if err != nil || co == nil {
		t.Fatal("expected company after expire")
	}
	if co.WalletQuotaRemain != 0 {
		t.Fatalf("after expire: wallet_quota_remain got %v want 0", co.WalletQuotaRemain)
	}

	// Lot should be expired.
	lots, err := st.Billing().ListActiveLotsFIFO(ctx, companyID, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(lots) != 0 {
		t.Fatalf("expected 0 active lots after expire, got %d", len(lots))
	}
}

func TestExpireMockLotsPreservesPaidLotBalance(t *testing.T) {
	t.Parallel()
	companyID := uuid.MustParse("00000000-0000-7000-0000-000000009204")
	_, st := testutil.NewTestStore(t)
	ctx := newLotTestCompany(t, st, companyID)

	// 1. Seed trial credit.
	trialQuota := int64(10000)
	if err := domainbilling.SeedTrialCredit(ctx, st, companyID, trialQuota, nil); err != nil {
		t.Fatal(err)
	}

	// 2. Add a paid lot (simulates real recharge before upgrade).
	ppu := domainbilling.DefaultQuotaPerUnit()
	paidAmount := float64(50)
	paidPoints := common.QuotaFromAmount(paidAmount, ppu)
	now := time.Now().UTC()
	paidOrder := store.RechargeOrder{
		ID: uuid.MustParse("00000000-0000-7000-0000-000000009204"), CompanyID: companyID, Amount: paidAmount,
		Currency: common.DefaultBillingCurrency, QuotaPerUnit: ppu, QuotaGranted: paidPoints,
		Source: store.RechargeSourceSelf, LotKind: store.LotKindPaid,
		Status: store.RechargeStatusConfirmed, DisplayOrderID: "ORD-9204",
		PaymentMethod: store.PaymentMethodAlipay, InvoiceStatus: store.InvoiceStatusNone,
		CreatedBy: contract.IDMemberAdmin, CreatedAt: now, UpdatedAt: now,
	}
	paidLot := domainbilling.BuildLot(paidOrder, common.DefaultBillingCurrency, store.LotKindPaid, paidOrder.Amount)
	if err := billinglot.CreditFromLot(ctx, st, paidOrder, paidLot, paidLot.QuotaGranted, nil); err != nil {
		t.Fatal(err)
	}

	// Wallet should have trial + paid.
	co, err := st.Company().GetByID(ctx, companyID)
	if err != nil || co == nil {
		t.Fatal("expected company")
	}
	expectedBefore := trialQuota + paidPoints
	if co.WalletQuotaRemain != expectedBefore {
		t.Fatalf("before expire: wallet_quota_remain got %v want %v", co.WalletQuotaRemain, expectedBefore)
	}

	// 3. Expire trial lots.
	if err := domainbilling.ExpireMockLots(ctx, st, companyID); err != nil {
		t.Fatal(err)
	}

	// Wallet should only have paid lot balance.
	co, err = st.Company().GetByID(ctx, companyID)
	if err != nil || co == nil {
		t.Fatal("expected company after expire")
	}
	if co.WalletQuotaRemain != paidPoints {
		t.Fatalf("after expire: wallet_quota_remain got %v want %v (paid lot only)", co.WalletQuotaRemain, paidPoints)
	}

	// Only paid lot should be active.
	lots, err := st.Billing().ListActiveLotsFIFO(ctx, companyID, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(lots) != 1 {
		t.Fatalf("expected 1 active lot (paid), got %d", len(lots))
	}
	if lots[0].LotKind != store.LotKindPaid {
		t.Fatalf("remaining lot kind: got %q want %q", lots[0].LotKind, store.LotKindPaid)
	}
}

func TestExpireMockLotsAfterPartialConsumption(t *testing.T) {
	t.Parallel()
	companyID := uuid.MustParse("00000000-0000-7000-0000-000000009205")
	_, st := testutil.NewTestStore(t)
	ctx := newLotTestCompany(t, st, companyID)

	// Seed trial credit.
	trialQuota := int64(10000)
	if err := domainbilling.SeedTrialCredit(ctx, st, companyID, trialQuota, nil); err != nil {
		t.Fatal(err)
	}

	// Consume part of the trial lot.
	consumed := int64(3000)
	_, err := billinglot.ConsumeLots(ctx, st, companyID, consumed)
	if err != nil {
		t.Fatal(err)
	}

	// Verify partial consumption.
	co, err := st.Company().GetByID(ctx, companyID)
	if err != nil || co == nil {
		t.Fatal("expected company")
	}
	if co.WalletQuotaRemain != trialQuota-consumed {
		t.Fatalf("after consume: wallet_quota_remain got %v want %v", co.WalletQuotaRemain, trialQuota-consumed)
	}

	// Expire trial lots.
	if err := domainbilling.ExpireMockLots(ctx, st, companyID); err != nil {
		t.Fatal(err)
	}

	// Wallet should be zero — trial lot (even partially consumed) is expired.
	co, err = st.Company().GetByID(ctx, companyID)
	if err != nil || co == nil {
		t.Fatal("expected company after expire")
	}
	if co.WalletQuotaRemain != 0 {
		t.Fatalf("after expire: wallet_quota_remain got %v want 0", co.WalletQuotaRemain)
	}
}

func TestExpireMockLotsIdempotent(t *testing.T) {
	t.Parallel()
	companyID := uuid.MustParse("00000000-0000-7000-0000-000000009206")
	_, st := testutil.NewTestStore(t)
	ctx := newLotTestCompany(t, st, companyID)

	trialQuota := int64(5000)
	if err := domainbilling.SeedTrialCredit(ctx, st, companyID, trialQuota, nil); err != nil {
		t.Fatal(err)
	}

	// First expire.
	if err := domainbilling.ExpireMockLots(ctx, st, companyID); err != nil {
		t.Fatal(err)
	}

	// Second expire should be a no-op, not error.
	if err := domainbilling.ExpireMockLots(ctx, st, companyID); err != nil {
		t.Fatalf("second expire should not error: %v", err)
	}

	co, err := st.Company().GetByID(ctx, companyID)
	if err != nil || co == nil {
		t.Fatal("expected company")
	}
	if co.WalletQuotaRemain != 0 {
		t.Fatalf("wallet_quota_remain should stay 0 after double expire, got %v", co.WalletQuotaRemain)
	}
}
