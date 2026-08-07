//go:build testhook

package redemption_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	"github.com/tokenjoy/backend/internal/domain/redemption"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/support/quota"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func seedRedemptionCode(t *testing.T, st store.Store, code string, faceValue float64, expiresAt time.Time) {
	t.Helper()
	rc := store.RedemptionCode{
		ID:        uuid.Must(uuid.NewV7()),
		Code:      code,
		BatchName: "test-batch",
		FaceValue: faceValue,
		Currency:  quota.DefaultBillingCurrency,
		Status:    store.RedemptionStatusUnused,
		ExpiresAt: expiresAt,
		CreatedBy: contract.IDMemberAdmin,
		Note:      "",
		CreatedAt: time.Now().UTC(),
	}
	if err := st.Redemption().BatchInsert(context.Background(), []store.RedemptionCode{rc}); err != nil {
		t.Fatal(err)
	}
}

func newBillingSvc(st store.Store) domainbilling.Service {
	return domainbilling.NewService(st, nil, nil)
}

func TestRedeemCodeSuccess(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()
	code := "TJ-AAAA-BBBB-CCCC"
	seedRedemptionCode(t, st, code, 100, time.Now().Add(24*time.Hour))

	svc := newBillingSvc(st)
	result, err := svc.RedeemCode(ctx, code, contract.IDMemberAdmin)
	if err != nil {
		t.Fatal(err)
	}
	if result.FaceValue != 100 {
		t.Fatalf("faceValue: got %v want 100", result.FaceValue)
	}
	if result.QuotaGranted != quota.MoneyToQuota(100, quota.DefaultQuotaPerUnit) {
		t.Fatalf("quotaGranted: got %v want %v", result.QuotaGranted, quota.MoneyToQuota(100, quota.DefaultQuotaPerUnit))
	}

	// Verify code is now marked used (via List, since GetCodeForUpdate requires tx).
	list, err := st.Redemption().List(ctx, store.RedemptionListFilter{Page: 1, PageSize: 100, Status: ptrStr(store.RedemptionStatusUsed)})
	if err != nil {
		t.Fatal(err)
	}
	var rc *store.RedemptionCode
	for i := range list.Items {
		if list.Items[i].Code == code {
			rc = &list.Items[i]
			break
		}
	}
	if rc == nil {
		t.Fatal("expected to find used code in list")
	}
	if rc.Status != store.RedemptionStatusUsed {
		t.Fatalf("status: got %q want %q", rc.Status, store.RedemptionStatusUsed)
	}
	if rc.RedeemedByCompany == nil || *rc.RedeemedByCompany != contract.DefaultCompanyID {
		t.Fatalf("redeemed_by_company: got %v want %v", rc.RedeemedByCompany, contract.DefaultCompanyID)
	}
}

func TestRedeemCodeAlreadyUsed(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()
	code := "TJ-USED-USED-USED"
	seedRedemptionCode(t, st, code, 50, time.Now().Add(24*time.Hour))

	svc := newBillingSvc(st)
	// First redeem succeeds.
	if _, err := svc.RedeemCode(ctx, code, contract.IDMemberAdmin); err != nil {
		t.Fatal(err)
	}
	// Second redeem fails.
	_, err := svc.RedeemCode(ctx, code, contract.IDMemberAdmin)
	if err == nil {
		t.Fatal("expected error for already-used code")
	}
	if !containsCode(err, "CODE_ALREADY_USED") {
		t.Fatalf("expected CODE_ALREADY_USED error, got: %v", err)
	}
}

func TestRedeemCodeExpired(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()
	code := "TJ-EXPD-EXPD-EXPD"
	seedRedemptionCode(t, st, code, 50, time.Now().Add(-1*time.Hour))

	svc := newBillingSvc(st)
	_, err := svc.RedeemCode(ctx, code, contract.IDMemberAdmin)
	if err == nil {
		t.Fatal("expected error for expired code")
	}
	if !containsCode(err, "CODE_EXPIRED") {
		t.Fatalf("expected CODE_EXPIRED error, got: %v", err)
	}
}

func TestRedeemCodeInvalid(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()

	svc := newBillingSvc(st)
	_, err := svc.RedeemCode(ctx, "TJ-NOPE-NOPE-NOPE", contract.IDMemberAdmin)
	if err == nil {
		t.Fatal("expected error for nonexistent code")
	}
	if !containsCode(err, "INVALID_REDEMPTION_CODE") {
		t.Fatalf("expected INVALID_REDEMPTION_CODE error, got: %v", err)
	}
}

func TestRedeemCodeNormalizesInput(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()
	code := "TJ-NORM-TEST-ABCD"
	seedRedemptionCode(t, st, code, 25, time.Now().Add(24*time.Hour))

	svc := newBillingSvc(st)
	// Input without dashes and lowercase should still work.
	result, err := svc.RedeemCode(ctx, "tjnormtestabcd", contract.IDMemberAdmin)
	if err != nil {
		t.Fatalf("expected normalization to match: %v", err)
	}
	if result.FaceValue != 25 {
		t.Fatalf("faceValue: got %v want 25", result.FaceValue)
	}
}

func TestGenerateBatchInserts(t *testing.T) {
	t.Parallel()
	_, st := testutil.NewTestStore(t)

	svc := redemption.NewService(st)
	result, err := svc.Generate(context.Background(), redemption.GenerateInput{
		BatchName:     "test-gen",
		FaceValue:     10,
		Quantity:      5,
		ExpiresInDays: 30,
		CreatedBy:     contract.IDMemberAdmin,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Quantity != 5 {
		t.Fatalf("quantity: got %v want 5", result.Quantity)
	}
	if result.FaceValue != 10 {
		t.Fatalf("faceValue: got %v want 10", result.FaceValue)
	}

	// Verify codes exist in DB.
	batch := "test-gen"
	list, err := st.Redemption().List(context.Background(), store.RedemptionListFilter{
		BatchName: &batch,
		Page:      1,
		PageSize:  20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if list.Total != 5 {
		t.Fatalf("list total: got %v want 5", list.Total)
	}
	for _, rc := range list.Items {
		if rc.FaceValue != 10 {
			t.Fatalf("item face_value: got %v want 10", rc.FaceValue)
		}
		if rc.Status != store.RedemptionStatusUnused {
			t.Fatalf("item status: got %q want %q", rc.Status, store.RedemptionStatusUnused)
		}
	}
}

// containsCode checks if a domain error message carries the given code.
func containsCode(err error, code string) bool {
	return err != nil && strings.Contains(err.Error(), code)
}

func ptrStr(s string) *string { return &s }
