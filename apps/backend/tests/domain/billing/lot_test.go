package billing_test

import (
	"context"
	"testing"
	"time"

	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	billinglot "github.com/tokenjoy/backend/internal/domain/billing/lot"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func newLotTestCompany(t *testing.T, st store.Store, companyID int64) context.Context {
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

func paidRechargeOrder(companyID int64, id string, amount float64, createdAt time.Time) store.RechargeOrder {
	ppu := domainbilling.DefaultPointsPerUnit()
	return store.RechargeOrder{
		ID: id, CompanyID: companyID, Amount: amount, Currency: common.DefaultBillingCurrency,
		PointsPerUnit: ppu, PointsGranted: domainbilling.PointsGrantedFromAmount(amount, ppu),
		Source: store.RechargeSourceSelf, LotKind: store.LotKindPaid,
		Status:         store.RechargeStatusConfirmed,
		DisplayOrderID: "ORD-" + id,
		PaymentMethod:  store.PaymentMethodAlipay,
		InvoiceStatus:  store.InvoiceStatusNone,
		CreatedBy:      "m-admin", CreatedAt: createdAt, UpdatedAt: createdAt,
	}
}

func TestCreditFromLotUpdatesWalletRemain(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()
	now := time.Now().UTC()
	ppu := domainbilling.DefaultPointsPerUnit()

	co, err := st.Company().GetByID(ctx, contract.DefaultCompanyID)
	if err != nil || co == nil {
		t.Fatal("expected default company")
	}
	before := co.WalletRemain

	key := "idem-wallet-credit"
	order := store.RechargeOrder{
		ID: "rch-wallet-1", CompanyID: contract.DefaultCompanyID, Amount: 50, Currency: common.DefaultBillingCurrency,
		PointsPerUnit: ppu, PointsGranted: domainbilling.PointsGrantedFromAmount(50, ppu),
		Source: store.RechargeSourceSelf, LotKind: store.LotKindPaid,
		IdempotencyKey: &key, Status: store.RechargeStatusConfirmed,
		DisplayOrderID: "ORD20260101130000",
		PaymentMethod:  store.PaymentMethodAlipay,
		InvoiceStatus:  store.InvoiceStatusNone,
		CreatedBy:      "m-admin", CreatedAt: now, UpdatedAt: now,
	}
	lot := domainbilling.BuildPaidLot(order, common.DefaultBillingCurrency, ppu)
	if err := billinglot.CreditFromLot(ctx, st, order, lot, lot.PointsGranted); err != nil {
		t.Fatal(err)
	}

	got, err := st.Company().GetByID(ctx, contract.DefaultCompanyID)
	if err != nil || got == nil {
		t.Fatal("expected company after credit")
	}
	want := before + lot.PointsGranted
	if got.WalletRemain != want {
		t.Fatalf("wallet_remain: got %v want %v", got.WalletRemain, want)
	}
}

func TestConsumeLotsDecrementsWalletRemain(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()
	now := time.Now().UTC()
	ppu := domainbilling.DefaultPointsPerUnit()
	grant := domainbilling.PointsGrantedFromAmount(10, ppu)

	order := store.RechargeOrder{
		ID: "rch-wallet-consume", CompanyID: contract.DefaultCompanyID, Amount: 10, Currency: common.DefaultBillingCurrency,
		PointsPerUnit: ppu, PointsGranted: grant,
		Source: store.RechargeSourceSelf, LotKind: store.LotKindPaid,
		Status:         store.RechargeStatusConfirmed,
		DisplayOrderID: "ORD20260101140000",
		PaymentMethod:  store.PaymentMethodAlipay,
		InvoiceStatus:  store.InvoiceStatusNone,
		CreatedBy:      "m-admin", CreatedAt: now, UpdatedAt: now,
	}
	lot := domainbilling.BuildPaidLot(order, common.DefaultBillingCurrency, ppu)
	if err := billinglot.CreditFromLot(ctx, st, order, lot, lot.PointsGranted); err != nil {
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
	want := afterCredit.WalletRemain - consume
	if afterConsume.WalletRemain != want {
		t.Fatalf("wallet_remain: got %v want %v", afterConsume.WalletRemain, want)
	}
}

func TestCreditFromLotSetsFIFOHeadOnFirstRecharge(t *testing.T) {
	t.Parallel()
	const companyID int64 = 9101
	_, st := testutil.NewTestStore(t)
	ctx := newLotTestCompany(t, st, companyID)
	now := time.Now().UTC()

	order := paidRechargeOrder(companyID, "rch-fifo-first", 20, now)
	lot := domainbilling.BuildPaidLot(order, common.DefaultBillingCurrency, domainbilling.DefaultPointsPerUnit())
	if err := billinglot.CreditFromLot(ctx, st, order, lot, lot.PointsGranted); err != nil {
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
	const companyID int64 = 9102
	_, st := testutil.NewTestStore(t)
	ctx := newLotTestCompany(t, st, companyID)
	base := time.Now().UTC()
	ppu := domainbilling.DefaultPointsPerUnit()

	orderA := paidRechargeOrder(companyID, "rch-fifo-a", 30, base)
	lotA := domainbilling.BuildPaidLot(orderA, common.DefaultBillingCurrency, ppu)
	if err := billinglot.CreditFromLot(ctx, st, orderA, lotA, lotA.PointsGranted); err != nil {
		t.Fatal(err)
	}

	orderB := paidRechargeOrder(companyID, "rch-fifo-b", 40, base.Add(time.Second))
	lotB := domainbilling.BuildPaidLot(orderB, common.DefaultBillingCurrency, ppu)
	if err := billinglot.CreditFromLot(ctx, st, orderB, lotB, lotB.PointsGranted); err != nil {
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
	const companyID int64 = 9103
	_, st := testutil.NewTestStore(t)
	ctx := newLotTestCompany(t, st, companyID)
	base := time.Now().UTC()
	ppu := domainbilling.DefaultPointsPerUnit()

	orderA := paidRechargeOrder(companyID, "rch-consume-a", 100, base)
	lotA := domainbilling.BuildPaidLot(orderA, common.DefaultBillingCurrency, ppu)
	if err := billinglot.CreditFromLot(ctx, st, orderA, lotA, lotA.PointsGranted); err != nil {
		t.Fatal(err)
	}

	orderB := paidRechargeOrder(companyID, "rch-consume-b", 100, base.Add(time.Second))
	lotB := domainbilling.BuildPaidLot(orderB, common.DefaultBillingCurrency, ppu)
	if err := billinglot.CreditFromLot(ctx, st, orderB, lotB, lotB.PointsGranted); err != nil {
		t.Fatal(err)
	}

	consume := lotA.PointsGranted / 4
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
	if segments[0].Points != consume {
		t.Fatalf("segment points: got %v want %v", segments[0].Points, consume)
	}

	gotA, err := st.Billing().GetLotByID(ctx, lotA.ID)
	if err != nil || gotA == nil {
		t.Fatal("expected lot A")
	}
	wantA := lotA.PointsGranted - consume
	if gotA.PointsRemaining != wantA {
		t.Fatalf("lot A remaining: got %v want %v", gotA.PointsRemaining, wantA)
	}

	gotB, err := st.Billing().GetLotByID(ctx, lotB.ID)
	if err != nil || gotB == nil {
		t.Fatal("expected lot B")
	}
	if gotB.PointsRemaining != lotB.PointsGranted {
		t.Fatalf("lot B should be untouched: got %v want %v", gotB.PointsRemaining, lotB.PointsGranted)
	}
}

func TestConsumeLotsMovesToNextLotAfterFirstExhausted(t *testing.T) {
	t.Parallel()
	const companyID int64 = 9104
	_, st := testutil.NewTestStore(t)
	ctx := newLotTestCompany(t, st, companyID)
	base := time.Now().UTC()
	ppu := domainbilling.DefaultPointsPerUnit()

	orderA := paidRechargeOrder(companyID, "rch-exhaust-a", 50, base)
	lotA := domainbilling.BuildPaidLot(orderA, common.DefaultBillingCurrency, ppu)
	if err := billinglot.CreditFromLot(ctx, st, orderA, lotA, lotA.PointsGranted); err != nil {
		t.Fatal(err)
	}

	orderB := paidRechargeOrder(companyID, "rch-exhaust-b", 80, base.Add(time.Second))
	lotB := domainbilling.BuildPaidLot(orderB, common.DefaultBillingCurrency, ppu)
	if err := billinglot.CreditFromLot(ctx, st, orderB, lotB, lotB.PointsGranted); err != nil {
		t.Fatal(err)
	}

	// Drain lot A completely, then take 10 from lot B.
	consume := lotA.PointsGranted + 10
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
	if segments[0].Points != lotA.PointsGranted {
		t.Fatalf("first segment points: got %v want %v", segments[0].Points, lotA.PointsGranted)
	}
	if segments[1].Points != 10 {
		t.Fatalf("second segment points: got %v want 10", segments[1].Points)
	}
	if result.OverdraftUsed {
		t.Fatal("expected no overdraft when paid lots cover the consume")
	}
}

func TestConsumeLotsExpandsOverdraftAndReportsDelta(t *testing.T) {
	t.Parallel()
	const companyID int64 = 9105
	_, st := testutil.NewTestStore(t)
	ctx := newLotTestCompany(t, st, companyID)
	now := time.Now().UTC()

	order := paidRechargeOrder(companyID, "rch-overdraft", 10, now)
	lot := domainbilling.BuildPaidLot(order, common.DefaultBillingCurrency, domainbilling.DefaultPointsPerUnit())
	if err := billinglot.CreditFromLot(ctx, st, order, lot, lot.PointsGranted); err != nil {
		t.Fatal(err)
	}

	extra := domainbilling.PointsGrantedFromAmount(3, domainbilling.DefaultPointsPerUnit())
	consume := lot.PointsGranted + extra
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
	if last.Points != extra {
		t.Fatalf("overdraft segment points: got %v want %v", last.Points, extra)
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
	if co.WalletRemain != 0 {
		t.Fatalf("wallet_remain after overdraft: got %v want 0", co.WalletRemain)
	}
}
