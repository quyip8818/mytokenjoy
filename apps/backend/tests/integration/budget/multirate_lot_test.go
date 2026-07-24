//go:build testhook && integration

package budget_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	billinglot "github.com/tokenjoy/backend/internal/domain/billing/lot"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/postgres"
	"github.com/tokenjoy/backend/tests/testutil"
)

// TestMultiRateRechargeAndConsume verifies the FIFO lot consumption scenario from the
// implementation guide §6:
//
//	lot1: QPU=500000, remaining=5M quota (= ¥10)
//	lot2: QPU=600000, remaining=6M quota (= ¥10)
//	consume 6M → lot1 drains 5M (display ¥10), lot2 takes 1M (display ¥1.6667)
//
// This exercises the key design property: each lot segment carries its own QuotaPerUnit
// snapshot, so DisplayAmount is computed per-lot, not globally.
func TestMultiRateRechargeAndConsume(t *testing.T) {
	t.Parallel()
	companyID := uuid.MustParse("00000000-0000-7000-0000-000000009201")
	_, st := testutil.NewTestStore(t)
	ctx := newCompany(t, st, companyID)
	base := time.Now().UTC()

	// --- Recharge lot 1: ¥10 at QPU=500000 → 5,000,000 quota ---
	qpu1 := int64(500_000)
	amount1 := 10.0
	quota1 := common.QuotaFromAmount(amount1, qpu1)
	if quota1 != 5_000_000 {
		t.Fatalf("lot1 quota: got %d want 5000000", quota1)
	}
	order1 := store.RechargeOrder{
		ID: uuid.MustParse("00000000-0000-7000-0000-000000009211"), CompanyID: companyID,
		Amount: amount1, Currency: common.DefaultBillingCurrency,
		QuotaPerUnit: qpu1, QuotaGranted: quota1,
		Source: store.RechargeSourceSelf, LotKind: store.LotKindPaid,
		Status: store.RechargeStatusConfirmed, DisplayOrderID: "ORD-rate-1",
		PaymentMethod: store.PaymentMethodAlipay, InvoiceStatus: store.InvoiceStatusNone,
		CreatedBy: uuid.Nil, CreatedAt: base, UpdatedAt: base,
	}
	lot1 := domainbilling.BuildLot(order1, common.DefaultBillingCurrency, store.LotKindPaid, order1.Amount)
	if err := billinglot.CreditFromLot(ctx, st, order1, lot1, lot1.QuotaGranted); err != nil {
		t.Fatal(err)
	}

	// --- Recharge lot 2: ¥10 at QPU=600000 → 6,000,000 quota ---
	qpu2 := int64(600_000)
	amount2 := 10.0
	quota2 := common.QuotaFromAmount(amount2, qpu2)
	if quota2 != 6_000_000 {
		t.Fatalf("lot2 quota: got %d want 6000000", quota2)
	}
	order2 := store.RechargeOrder{
		ID: uuid.MustParse("00000000-0000-7000-0000-000000009212"), CompanyID: companyID,
		Amount: amount2, Currency: common.DefaultBillingCurrency,
		QuotaPerUnit: qpu2, QuotaGranted: quota2,
		Source: store.RechargeSourceSelf, LotKind: store.LotKindPaid,
		Status: store.RechargeStatusConfirmed, DisplayOrderID: "ORD-rate-2",
		PaymentMethod: store.PaymentMethodAlipay, InvoiceStatus: store.InvoiceStatusNone,
		CreatedBy: uuid.Nil, CreatedAt: base.Add(time.Second), UpdatedAt: base.Add(time.Second),
	}
	lot2 := domainbilling.BuildLot(order2, common.DefaultBillingCurrency, store.LotKindPaid, order2.Amount)
	if err := billinglot.CreditFromLot(ctx, st, order2, lot2, lot2.QuotaGranted); err != nil {
		t.Fatal(err)
	}

	// Verify wallet after both recharges.
	co, err := st.Company().GetByID(ctx, companyID)
	if err != nil || co == nil {
		t.Fatal("expected company")
	}
	wantWallet := quota1 + quota2 // 11,000,000
	if co.WalletRemainQuota != wantWallet {
		t.Fatalf("wallet after recharge: got %d want %d", co.WalletRemainQuota, wantWallet)
	}

	// --- Consume 6,000,000 quota (drains lot1=5M, then 1M from lot2) ---
	consume := int64(6_000_000)
	result, err := billinglot.ConsumeLots(ctx, st, companyID, consume)
	if err != nil {
		t.Fatal(err)
	}

	// Expect 2 segments.
	if len(result.Segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(result.Segments))
	}

	seg1 := result.Segments[0]
	seg2 := result.Segments[1]

	// Segment 1: all of lot1 (5M).
	if seg1.LotID != lot1.ID {
		t.Fatalf("seg1 lotID: got %s want %s", seg1.LotID, lot1.ID)
	}
	if seg1.Quota != 5_000_000 {
		t.Fatalf("seg1 quota: got %d want 5000000", seg1.Quota)
	}
	if seg1.QuotaPerUnit != qpu1 {
		t.Fatalf("seg1 QPU: got %d want %d", seg1.QuotaPerUnit, qpu1)
	}
	// DisplayAmount = 5000000 / 500000 = 10.0
	if seg1.DisplayAmount != 10.0 {
		t.Fatalf("seg1 display: got %v want 10.0", seg1.DisplayAmount)
	}

	// Segment 2: 1M from lot2.
	if seg2.LotID != lot2.ID {
		t.Fatalf("seg2 lotID: got %s want %s", seg2.LotID, lot2.ID)
	}
	if seg2.Quota != 1_000_000 {
		t.Fatalf("seg2 quota: got %d want 1000000", seg2.Quota)
	}
	if seg2.QuotaPerUnit != qpu2 {
		t.Fatalf("seg2 QPU: got %d want %d", seg2.QuotaPerUnit, qpu2)
	}
	// DisplayAmount = 1000000 / 600000 ≈ 1.6667
	wantDisplay2 := float64(1_000_000) / float64(600_000)
	if math.Abs(seg2.DisplayAmount-wantDisplay2) > 0.0001 {
		t.Fatalf("seg2 display: got %v want ≈%v", seg2.DisplayAmount, wantDisplay2)
	}

	// Verify no overdraft.
	if result.OverdraftUsed {
		t.Fatal("expected no overdraft")
	}

	// Verify wallet decremented.
	co, err = st.Company().GetByID(ctx, companyID)
	if err != nil || co == nil {
		t.Fatal("expected company after consume")
	}
	if co.WalletRemainQuota != wantWallet-consume {
		t.Fatalf("wallet after consume: got %d want %d", co.WalletRemainQuota, wantWallet-consume)
	}

	// Verify lot1 is exhausted, lot2 has 5M remaining.
	gotLot1, err := st.Billing().GetLotByID(ctx, lot1.ID)
	if err != nil || gotLot1 == nil {
		t.Fatal("expected lot1")
	}
	if gotLot1.QuotaRemaining != 0 {
		t.Fatalf("lot1 remaining: got %d want 0", gotLot1.QuotaRemaining)
	}
	if gotLot1.Status != store.LotStatusExhausted {
		t.Fatalf("lot1 status: got %s want exhausted", gotLot1.Status)
	}

	gotLot2, err := st.Billing().GetLotByID(ctx, lot2.ID)
	if err != nil || gotLot2 == nil {
		t.Fatal("expected lot2")
	}
	if gotLot2.QuotaRemaining != 5_000_000 {
		t.Fatalf("lot2 remaining: got %d want 5000000", gotLot2.QuotaRemaining)
	}

	// --- Verify LedgerSegments carry per-lot display amounts ---
	baseEntry := types.UsageLedgerEntry{
		ID:        uuid.Must(uuid.NewV7()),
		Amount:    consume,
		EventType: types.EventTypeCallSettled,
	}
	ledgerEntries := billinglot.LedgerSegmentsFromEntry(baseEntry, result.Segments)
	if len(ledgerEntries) != 2 {
		t.Fatalf("expected 2 ledger entries, got %d", len(ledgerEntries))
	}
	if ledgerEntries[0].Amount != 5_000_000 {
		t.Fatalf("ledger[0].Amount: got %d want 5000000", ledgerEntries[0].Amount)
	}
	if ledgerEntries[0].DisplayAmount != 10.0 {
		t.Fatalf("ledger[0].DisplayAmount: got %v want 10.0", ledgerEntries[0].DisplayAmount)
	}
	if ledgerEntries[1].Amount != 1_000_000 {
		t.Fatalf("ledger[1].Amount: got %d want 1000000", ledgerEntries[1].Amount)
	}
	if math.Abs(ledgerEntries[1].DisplayAmount-wantDisplay2) > 0.0001 {
		t.Fatalf("ledger[1].DisplayAmount: got %v want ≈%v", ledgerEntries[1].DisplayAmount, wantDisplay2)
	}
}

// TestGiftLotUsesCompanyQPUForDisplay verifies that gift lots (AmountDisplay=0)
// still produce meaningful DisplayAmount during consumption using the lot's QPU snapshot.
func TestGiftLotUsesCompanyQPUForDisplay(t *testing.T) {
	t.Parallel()
	companyID := uuid.MustParse("00000000-0000-7000-0000-000000009202")
	_, st := testutil.NewTestStore(t)
	ctx := newCompany(t, st, companyID)
	now := time.Now().UTC()

	qpu := int64(500_000)
	giftQuota := int64(1_000_000) // equivalent to ¥2
	order := store.RechargeOrder{
		ID: uuid.MustParse("00000000-0000-7000-0000-000000009221"), CompanyID: companyID,
		Amount: 0, Currency: common.DefaultBillingCurrency,
		QuotaPerUnit: qpu, QuotaGranted: giftQuota,
		Source: store.RechargeSourceGift, LotKind: store.LotKindGift,
		Status:    store.RechargeStatusConfirmed,
		CreatedBy: uuid.Nil, CreatedAt: now, UpdatedAt: now,
	}
	lot := domainbilling.BuildLot(order, common.DefaultBillingCurrency, store.LotKindGift, 0)
	if err := billinglot.CreditFromLot(ctx, st, order, lot, lot.QuotaGranted); err != nil {
		t.Fatal(err)
	}

	// Consume all 1M.
	result, err := billinglot.ConsumeLots(ctx, st, companyID, giftQuota)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(result.Segments))
	}
	seg := result.Segments[0]
	if seg.Quota != giftQuota {
		t.Fatalf("segment quota: got %d want %d", seg.Quota, giftQuota)
	}
	// Gift lot DisplayAmount = 1000000 / 500000 = ¥2.00 (equivalent value, not "paid")
	if seg.DisplayAmount != 2.0 {
		t.Fatalf("gift segment display: got %v want 2.0", seg.DisplayAmount)
	}
}

// TestOverdraftWithMultiRateLots verifies that when paid lots are exhausted,
// the overdraft segment still produces correct display amounts.
func TestOverdraftWithMultiRateLots(t *testing.T) {
	t.Parallel()
	companyID := uuid.MustParse("00000000-0000-7000-0000-000000009203")
	_, st := testutil.NewTestStore(t)
	ctx := newCompany(t, st, companyID)
	now := time.Now().UTC()

	// Small paid lot: ¥1 at QPU=500000 → 500,000 quota.
	qpu := int64(500_000)
	order := store.RechargeOrder{
		ID: uuid.MustParse("00000000-0000-7000-0000-000000009231"), CompanyID: companyID,
		Amount: 1, Currency: common.DefaultBillingCurrency,
		QuotaPerUnit: qpu, QuotaGranted: 500_000,
		Source: store.RechargeSourceSelf, LotKind: store.LotKindPaid,
		Status: store.RechargeStatusConfirmed, DisplayOrderID: "ORD-od-1",
		PaymentMethod: store.PaymentMethodAlipay, InvoiceStatus: store.InvoiceStatusNone,
		CreatedBy: uuid.Nil, CreatedAt: now, UpdatedAt: now,
	}
	lot := domainbilling.BuildLot(order, common.DefaultBillingCurrency, store.LotKindPaid, order.Amount)
	if err := billinglot.CreditFromLot(ctx, st, order, lot, lot.QuotaGranted); err != nil {
		t.Fatal(err)
	}

	// Consume 800,000 (exceeds paid lot by 300,000).
	consume := int64(800_000)
	result, err := billinglot.ConsumeLots(ctx, st, companyID, consume)
	if err != nil {
		t.Fatal(err)
	}
	if !result.OverdraftUsed {
		t.Fatal("expected overdraft")
	}
	if result.OverdraftDelta != 300_000 {
		t.Fatalf("overdraft delta: got %d want 300000", result.OverdraftDelta)
	}
	if len(result.Segments) != 2 {
		t.Fatalf("expected 2 segments (paid + overdraft), got %d", len(result.Segments))
	}

	// Paid segment: 500,000 quota, display = ¥1.
	if result.Segments[0].Quota != 500_000 {
		t.Fatalf("paid segment quota: got %d want 500000", result.Segments[0].Quota)
	}
	if result.Segments[0].DisplayAmount != 1.0 {
		t.Fatalf("paid segment display: got %v want 1.0", result.Segments[0].DisplayAmount)
	}

	// Overdraft segment: 300,000 quota.
	if result.Segments[1].Quota != 300_000 {
		t.Fatalf("overdraft segment quota: got %d want 300000", result.Segments[1].Quota)
	}

	// Wallet should be 0 after overdraft.
	co, err := st.Company().GetByID(ctx, companyID)
	if err != nil || co == nil {
		t.Fatal("expected company")
	}
	if co.WalletRemainQuota != 0 {
		t.Fatalf("wallet after overdraft: got %d want 0", co.WalletRemainQuota)
	}
}

// TestCompanyQPUSwitchMidLifecycle simulates a real scenario where a company's
// effective QPU changes between two recharges (e.g. currencies table updated by admin).
// Lot1 is created at QPU=500000, then the currency rate changes to 600000,
// and lot2 is created at QPU=600000. Consumption crosses both lots with different
// display conversions per segment.
func TestCompanyQPUSwitchMidLifecycle(t *testing.T) {
	t.Parallel()
	companyID := uuid.MustParse("00000000-0000-7000-0000-000000009204")
	_, st := testutil.NewTestStore(t)
	ctx := newCompany(t, st, companyID)
	base := time.Now().UTC()
	pool := postgres.MainPool(st)

	// --- Phase 1: QPU = 500000 (default seeded value) ---
	qpu1 := int64(500_000)

	// Recharge ¥20 → 10,000,000 quota at QPU=500000.
	amount1 := 20.0
	quota1 := common.QuotaFromAmount(amount1, qpu1)
	order1 := store.RechargeOrder{
		ID: uuid.MustParse("00000000-0000-7000-0000-000000009241"), CompanyID: companyID,
		Amount: amount1, Currency: common.DefaultBillingCurrency,
		QuotaPerUnit: qpu1, QuotaGranted: quota1,
		Source: store.RechargeSourceSelf, LotKind: store.LotKindPaid,
		Status: store.RechargeStatusConfirmed, DisplayOrderID: "ORD-switch-1",
		PaymentMethod: store.PaymentMethodAlipay, InvoiceStatus: store.InvoiceStatusNone,
		CreatedBy: uuid.Nil, CreatedAt: base, UpdatedAt: base,
	}
	lot1 := domainbilling.BuildLot(order1, common.DefaultBillingCurrency, store.LotKindPaid, order1.Amount)
	if err := billinglot.CreditFromLot(ctx, st, order1, lot1, lot1.QuotaGranted); err != nil {
		t.Fatal(err)
	}

	// --- Admin changes currency rate to 600000 ---
	qpu2 := int64(600_000)
	if _, err := pool.Exec(ctx, `
		UPDATE currencies SET quota_per_unit = $1 WHERE currency = $2
	`, qpu2, common.DefaultBillingCurrency); err != nil {
		t.Fatalf("update currencies QPU: %v", err)
	}
	// Restore at end regardless of outcome.
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `
			UPDATE currencies SET quota_per_unit = $1 WHERE currency = $2
		`, common.DefaultQuotaPerUnit, common.DefaultBillingCurrency)
	})

	// --- Phase 2: QPU = 600000 ---
	// Recharge ¥20 → 12,000,000 quota at QPU=600000.
	amount2 := 20.0
	quota2 := common.QuotaFromAmount(amount2, qpu2)
	if quota2 != 12_000_000 {
		t.Fatalf("lot2 quota: got %d want 12000000", quota2)
	}
	order2 := store.RechargeOrder{
		ID: uuid.MustParse("00000000-0000-7000-0000-000000009242"), CompanyID: companyID,
		Amount: amount2, Currency: common.DefaultBillingCurrency,
		QuotaPerUnit: qpu2, QuotaGranted: quota2,
		Source: store.RechargeSourceSelf, LotKind: store.LotKindPaid,
		Status: store.RechargeStatusConfirmed, DisplayOrderID: "ORD-switch-2",
		PaymentMethod: store.PaymentMethodAlipay, InvoiceStatus: store.InvoiceStatusNone,
		CreatedBy: uuid.Nil, CreatedAt: base.Add(time.Second), UpdatedAt: base.Add(time.Second),
	}
	lot2 := domainbilling.BuildLot(order2, common.DefaultBillingCurrency, store.LotKindPaid, order2.Amount)
	if err := billinglot.CreditFromLot(ctx, st, order2, lot2, lot2.QuotaGranted); err != nil {
		t.Fatal(err)
	}

	// Verify wallet = lot1 + lot2 = 10M + 12M = 22M.
	co, err := st.Company().GetByID(ctx, companyID)
	if err != nil || co == nil {
		t.Fatal("expected company")
	}
	if co.WalletRemainQuota != quota1+quota2 {
		t.Fatalf("wallet: got %d want %d", co.WalletRemainQuota, quota1+quota2)
	}

	// --- Consume 11,000,000 quota (drains lot1=10M, then 1M from lot2) ---
	consume := int64(11_000_000)
	result, err := billinglot.ConsumeLots(ctx, st, companyID, consume)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(result.Segments))
	}

	seg1 := result.Segments[0]
	seg2 := result.Segments[1]

	// Segment 1: drains lot1 entirely (10M at QPU=500000 → display ¥20).
	if seg1.Quota != quota1 {
		t.Fatalf("seg1 quota: got %d want %d", seg1.Quota, quota1)
	}
	if seg1.QuotaPerUnit != qpu1 {
		t.Fatalf("seg1 QPU snapshot: got %d want %d (original rate)", seg1.QuotaPerUnit, qpu1)
	}
	wantDisplay1 := float64(quota1) / float64(qpu1) // 10M/500000 = 20.0
	if seg1.DisplayAmount != wantDisplay1 {
		t.Fatalf("seg1 display: got %v want %v", seg1.DisplayAmount, wantDisplay1)
	}

	// Segment 2: 1M from lot2 (QPU=600000 → display ≈ ¥1.6667).
	if seg2.Quota != 1_000_000 {
		t.Fatalf("seg2 quota: got %d want 1000000", seg2.Quota)
	}
	if seg2.QuotaPerUnit != qpu2 {
		t.Fatalf("seg2 QPU snapshot: got %d want %d (new rate)", seg2.QuotaPerUnit, qpu2)
	}
	wantDisplay2 := float64(1_000_000) / float64(qpu2)
	if math.Abs(seg2.DisplayAmount-wantDisplay2) > 0.0001 {
		t.Fatalf("seg2 display: got %v want ≈%v", seg2.DisplayAmount, wantDisplay2)
	}

	// Key invariant: total display cost ≠ simple quota/currentQPU.
	// If you naively computed 11M / 600000 = ¥18.33 that would be WRONG.
	// Correct: ¥20 + ¥1.6667 = ¥21.6667.
	totalDisplay := seg1.DisplayAmount + seg2.DisplayAmount
	naiveDisplay := float64(consume) / float64(qpu2)
	if math.Abs(totalDisplay-naiveDisplay) < 1 {
		t.Fatalf("total display should NOT match naive single-rate calculation — got %v ≈ %v", totalDisplay, naiveDisplay)
	}
	wantTotal := wantDisplay1 + wantDisplay2 // 20 + 1.6667
	if math.Abs(totalDisplay-wantTotal) > 0.0001 {
		t.Fatalf("total display: got %v want ≈%v", totalDisplay, wantTotal)
	}

	// Verify lot states.
	gotLot1, _ := st.Billing().GetLotByID(ctx, lot1.ID)
	if gotLot1.QuotaRemaining != 0 || gotLot1.Status != store.LotStatusExhausted {
		t.Fatalf("lot1 should be exhausted: remaining=%d status=%s", gotLot1.QuotaRemaining, gotLot1.Status)
	}
	gotLot2, _ := st.Billing().GetLotByID(ctx, lot2.ID)
	if gotLot2.QuotaRemaining != quota2-1_000_000 {
		t.Fatalf("lot2 remaining: got %d want %d", gotLot2.QuotaRemaining, quota2-1_000_000)
	}
}

func newCompany(t *testing.T, st store.Store, companyID uuid.UUID) context.Context {
	t.Helper()
	ctx := testutil.CtxForCompany(companyID)
	now := time.Now().UTC()
	co := store.Company{
		ID: companyID, Name: "Multi-Rate Test Co",
		Status: store.CompanyStatusActive, CreatedAt: now, UpdatedAt: now,
	}
	if err := st.Company().Create(ctx, co); err != nil {
		t.Fatal(err)
	}
	return ctx
}
