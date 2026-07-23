package billing_test

import (
	"context"
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

func newLotTestCompany(t *testing.T, st store.Store, companyID uuid.UUID) context.Context {
	t.Helper()
	ctx := testutil.CtxForCompany(companyID)
	now := time.Now().UTC()
	co := store.Company{
		ID: companyID, Name: "Lot Test Co",
		Status: store.CompanyStatusActive, CreatedAt: now, UpdatedAt: now,
	}
	if err := st.Company().Create(ctx, co); err != nil {
		t.Fatal(err)
	}
	return ctx
}

func paidRechargeOrder(companyID uuid.UUID, id uuid.UUID, amount float64, createdAt time.Time) store.RechargeOrder {
	ppu := domainbilling.DefaultQuotaPerUnit()
	return store.RechargeOrder{
		ID: id, CompanyID: companyID, Amount: amount, Currency: common.DefaultBillingCurrency,
		QuotaPerUnit: ppu, QuotaGranted: common.QuotaFromAmount(amount, ppu),
		Source: store.RechargeSourceSelf, LotKind: store.LotKindPaid,
		Status:         store.RechargeStatusConfirmed,
		DisplayOrderID: "ORD-" + id.String(),
		PaymentMethod:  store.PaymentMethodAlipay,
		InvoiceStatus:  store.InvoiceStatusNone,
		CreatedBy:      contract.IDMemberAdmin, CreatedAt: createdAt, UpdatedAt: createdAt,
	}
}

func TestCreditFromLotUpdatesWalletQuotaRemain(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()
	now := time.Now().UTC()
	ppu := domainbilling.DefaultQuotaPerUnit()

	co, err := st.Company().GetByID(ctx, contract.DefaultCompanyID)
	if err != nil || co == nil {
		t.Fatal("expected default company")
	}
	before := co.WalletQuotaRemain

	key := "idem-wallet-credit"
	order := store.RechargeOrder{
		ID: uuid.MustParse("00000000-0000-7000-0000-000000002001"), CompanyID: contract.DefaultCompanyID, Amount: 50, Currency: common.DefaultBillingCurrency,
		QuotaPerUnit: ppu, QuotaGranted: common.QuotaFromAmount(50, ppu),
		Source: store.RechargeSourceSelf, LotKind: store.LotKindPaid,
		IdempotencyKey: &key, Status: store.RechargeStatusConfirmed,
		DisplayOrderID: "ORD20260101130000",
		PaymentMethod:  store.PaymentMethodAlipay,
		InvoiceStatus:  store.InvoiceStatusNone,
		CreatedBy:      contract.IDMemberAdmin, CreatedAt: now, UpdatedAt: now,
	}
	lot := domainbilling.BuildLot(order, common.DefaultBillingCurrency, store.LotKindPaid, order.Amount)
	if err := billinglot.CreditFromLot(ctx, st, order, lot, lot.QuotaGranted); err != nil {
		t.Fatal(err)
	}

	got, err := st.Company().GetByID(ctx, contract.DefaultCompanyID)
	if err != nil || got == nil {
		t.Fatal("expected company after credit")
	}
	want := before + lot.QuotaGranted
	if got.WalletQuotaRemain != want {
		t.Fatalf("wallet_quota_remain: got %v want %v", got.WalletQuotaRemain, want)
	}
}

func TestConsumeLotsDecrementsWalletQuotaRemain(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()
	now := time.Now().UTC()
	ppu := domainbilling.DefaultQuotaPerUnit()
	grant := common.QuotaFromAmount(10, ppu)

	order := store.RechargeOrder{
		ID: uuid.MustParse("00000000-0000-7000-0000-000000002002"), CompanyID: contract.DefaultCompanyID, Amount: 10, Currency: common.DefaultBillingCurrency,
		QuotaPerUnit: ppu, QuotaGranted: grant,
		Source: store.RechargeSourceSelf, LotKind: store.LotKindPaid,
		Status:         store.RechargeStatusConfirmed,
		DisplayOrderID: "ORD20260101140000",
		PaymentMethod:  store.PaymentMethodAlipay,
		InvoiceStatus:  store.InvoiceStatusNone,
		CreatedBy:      contract.IDMemberAdmin, CreatedAt: now, UpdatedAt: now,
	}
	lot := domainbilling.BuildLot(order, common.DefaultBillingCurrency, store.LotKindPaid, order.Amount)
	if err := billinglot.CreditFromLot(ctx, st, order, lot, lot.QuotaGranted); err != nil {
		t.Fatal(err)
	}
	afterCredit, err := st.Company().GetByID(ctx, contract.DefaultCompanyID)
	if err != nil || afterCredit == nil {
		t.Fatal("expected company after credit")
	}

	consume := grant / 4
	result, err := billinglot.ConsumeLots(ctx, st, contract.DefaultCompanyID, consume)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Segments) == 0 {
		t.Fatal("expected lot segments")
	}
	if result.OverdraftUsed {
		t.Fatal("expected no overdraft for partial consume within paid balance")
	}
	if result.OverdraftDelta != 0 {
		t.Fatalf("overdraft delta: got %v want 0", result.OverdraftDelta)
	}

	afterConsume, err := st.Company().GetByID(ctx, contract.DefaultCompanyID)
	if err != nil || afterConsume == nil {
		t.Fatal("expected company after consume")
	}
	want := afterCredit.WalletQuotaRemain - consume
	if afterConsume.WalletQuotaRemain != want {
		t.Fatalf("wallet_quota_remain: got %v want %v", afterConsume.WalletQuotaRemain, want)
	}
}

func TestCreditFromLotSetsFIFOHeadOnFirstRecharge(t *testing.T) {
	t.Parallel()
	companyID := uuid.MustParse("00000000-0000-7000-0000-000000009101")
	_, st := testutil.NewTestStore(t)
	ctx := newLotTestCompany(t, st, companyID)
	now := time.Now().UTC()

	order := paidRechargeOrder(companyID, uuid.MustParse("00000000-0000-7000-0000-000000001001"), 20, now)
	lot := domainbilling.BuildLot(order, common.DefaultBillingCurrency, store.LotKindPaid, order.Amount)
	if err := billinglot.CreditFromLot(ctx, st, order, lot, lot.QuotaGranted); err != nil {
		t.Fatal(err)
	}

	co, err := st.Company().GetByID(ctx, companyID)
	if err != nil || co == nil {
		t.Fatal("expected company")
	}
	if co.FIFOHeadLotID == nil || *co.FIFOHeadLotID != lot.ID {
		t.Fatalf("fifo head: got %v want %q", co.FIFOHeadLotID, lot.ID)
	}
}

func TestCreditFromLotPreservesFIFOHeadOnSecondRecharge(t *testing.T) {
	t.Parallel()
	companyID := uuid.MustParse("00000000-0000-7000-0000-000000009102")
	_, st := testutil.NewTestStore(t)
	ctx := newLotTestCompany(t, st, companyID)
	base := time.Now().UTC()

	orderA := paidRechargeOrder(companyID, uuid.MustParse("00000000-0000-7000-0000-000000001002"), 30, base)
	lotA := domainbilling.BuildLot(orderA, common.DefaultBillingCurrency, store.LotKindPaid, orderA.Amount)
	if err := billinglot.CreditFromLot(ctx, st, orderA, lotA, lotA.QuotaGranted); err != nil {
		t.Fatal(err)
	}

	orderB := paidRechargeOrder(companyID, uuid.MustParse("00000000-0000-7000-0000-000000001003"), 40, base.Add(time.Second))
	lotB := domainbilling.BuildLot(orderB, common.DefaultBillingCurrency, store.LotKindPaid, orderB.Amount)
	if err := billinglot.CreditFromLot(ctx, st, orderB, lotB, lotB.QuotaGranted); err != nil {
		t.Fatal(err)
	}

	co, err := st.Company().GetByID(ctx, companyID)
	if err != nil || co == nil {
		t.Fatal("expected company")
	}
	if co.FIFOHeadLotID == nil || *co.FIFOHeadLotID != lotA.ID {
		t.Fatalf("fifo head should stay on first lot: got %v want %q", co.FIFOHeadLotID, lotA.ID)
	}
	if *co.FIFOHeadLotID == lotB.ID {
		t.Fatalf("second lot must not become fifo head")
	}
}

func TestConsumeLotsDepletesOlderLotFirst(t *testing.T) {
	t.Parallel()
	companyID := uuid.MustParse("00000000-0000-7000-0000-000000009103")
	_, st := testutil.NewTestStore(t)
	ctx := newLotTestCompany(t, st, companyID)
	base := time.Now().UTC()

	orderA := paidRechargeOrder(companyID, uuid.MustParse("00000000-0000-7000-0000-000000001004"), 100, base)
	lotA := domainbilling.BuildLot(orderA, common.DefaultBillingCurrency, store.LotKindPaid, orderA.Amount)
	if err := billinglot.CreditFromLot(ctx, st, orderA, lotA, lotA.QuotaGranted); err != nil {
		t.Fatal(err)
	}

	orderB := paidRechargeOrder(companyID, uuid.MustParse("00000000-0000-7000-0000-000000001005"), 100, base.Add(time.Second))
	lotB := domainbilling.BuildLot(orderB, common.DefaultBillingCurrency, store.LotKindPaid, orderB.Amount)
	if err := billinglot.CreditFromLot(ctx, st, orderB, lotB, lotB.QuotaGranted); err != nil {
		t.Fatal(err)
	}

	consume := lotA.QuotaGranted / 4
	result, err := billinglot.ConsumeLots(ctx, st, companyID, consume)
	if err != nil {
		t.Fatal(err)
	}
	segments := result.Segments
	if len(segments) != 1 {
		t.Fatalf("expected single segment, got %d", len(segments))
	}
	if segments[0].LotID != lotA.ID {
		t.Fatalf("expected consumption from lot A %q, got %q", lotA.ID, segments[0].LotID)
	}
	if segments[0].Quota != consume {
		t.Fatalf("segment points: got %v want %v", segments[0].Quota, consume)
	}

	gotA, err := st.Billing().GetLotByID(ctx, lotA.ID)
	if err != nil || gotA == nil {
		t.Fatal("expected lot A")
	}
	wantA := lotA.QuotaGranted - consume
	if gotA.QuotaRemaining != wantA {
		t.Fatalf("lot A remaining: got %v want %v", gotA.QuotaRemaining, wantA)
	}

	gotB, err := st.Billing().GetLotByID(ctx, lotB.ID)
	if err != nil || gotB == nil {
		t.Fatal("expected lot B")
	}
	if gotB.QuotaRemaining != lotB.QuotaGranted {
		t.Fatalf("lot B should be untouched: got %v want %v", gotB.QuotaRemaining, lotB.QuotaGranted)
	}
}

func TestConsumeLotsMovesToNextLotAfterFirstExhausted(t *testing.T) {
	t.Parallel()
	companyID := uuid.MustParse("00000000-0000-7000-0000-000000009104")
	_, st := testutil.NewTestStore(t)
	ctx := newLotTestCompany(t, st, companyID)
	base := time.Now().UTC()

	orderA := paidRechargeOrder(companyID, uuid.MustParse("00000000-0000-7000-0000-000000001006"), 50, base)
	lotA := domainbilling.BuildLot(orderA, common.DefaultBillingCurrency, store.LotKindPaid, orderA.Amount)
	if err := billinglot.CreditFromLot(ctx, st, orderA, lotA, lotA.QuotaGranted); err != nil {
		t.Fatal(err)
	}

	orderB := paidRechargeOrder(companyID, uuid.MustParse("00000000-0000-7000-0000-000000001007"), 80, base.Add(time.Second))
	lotB := domainbilling.BuildLot(orderB, common.DefaultBillingCurrency, store.LotKindPaid, orderB.Amount)
	if err := billinglot.CreditFromLot(ctx, st, orderB, lotB, lotB.QuotaGranted); err != nil {
		t.Fatal(err)
	}

	// Drain lot A completely, then take 10 from lot B.
	consume := lotA.QuotaGranted + 10
	result, err := billinglot.ConsumeLots(ctx, st, companyID, consume)
	if err != nil {
		t.Fatal(err)
	}
	segments := result.Segments
	if len(segments) != 2 {
		t.Fatalf("expected two segments, got %d", len(segments))
	}
	if segments[0].LotID != lotA.ID || segments[1].LotID != lotB.ID {
		t.Fatalf("segment lot order: got [%q, %q] want [%q, %q]",
			segments[0].LotID, segments[1].LotID, lotA.ID, lotB.ID)
	}
	if segments[0].Quota != lotA.QuotaGranted {
		t.Fatalf("first segment points: got %v want %v", segments[0].Quota, lotA.QuotaGranted)
	}
	if segments[1].Quota != 10 {
		t.Fatalf("second segment points: got %v want 10", segments[1].Quota)
	}
	if result.OverdraftUsed {
		t.Fatal("expected no overdraft when paid lots cover the consume")
	}
}

func TestConsumeLotsExpandsOverdraftAndReportsDelta(t *testing.T) {
	t.Parallel()
	companyID := uuid.MustParse("00000000-0000-7000-0000-000000009105")
	_, st := testutil.NewTestStore(t)
	ctx := newLotTestCompany(t, st, companyID)
	now := time.Now().UTC()

	order := paidRechargeOrder(companyID, uuid.MustParse("00000000-0000-7000-0000-000000001008"), 10, now)
	lot := domainbilling.BuildLot(order, common.DefaultBillingCurrency, store.LotKindPaid, order.Amount)
	if err := billinglot.CreditFromLot(ctx, st, order, lot, lot.QuotaGranted); err != nil {
		t.Fatal(err)
	}

	extra := common.QuotaFromAmount(3, domainbilling.DefaultQuotaPerUnit())
	consume := lot.QuotaGranted + extra
	result, err := billinglot.ConsumeLots(ctx, st, companyID, consume)
	if err != nil {
		t.Fatal(err)
	}
	if !result.OverdraftUsed {
		t.Fatal("expected overdraft when consume exceeds paid lots")
	}
	if result.OverdraftDelta != extra {
		t.Fatalf("overdraft delta: got %v want %v", result.OverdraftDelta, extra)
	}
	if len(result.Segments) < 2 {
		t.Fatalf("expected paid + overdraft segments, got %d", len(result.Segments))
	}
	last := result.Segments[len(result.Segments)-1]
	if last.Quota != extra {
		t.Fatalf("overdraft segment points: got %v want %v", last.Quota, extra)
	}
	od, err := st.Billing().GetLotByID(ctx, last.LotID)
	if err != nil || od == nil {
		t.Fatal("expected overdraft lot")
	}
	if od.LotKind != store.LotKindOverdraft {
		t.Fatalf("lot kind: got %q want %q", od.LotKind, store.LotKindOverdraft)
	}

	co, err := st.Company().GetByID(ctx, companyID)
	if err != nil || co == nil {
		t.Fatal("expected company")
	}
	if co.WalletQuotaRemain != 0 {
		t.Fatalf("wallet_quota_remain after overdraft: got %v want 0", co.WalletQuotaRemain)
	}
}
