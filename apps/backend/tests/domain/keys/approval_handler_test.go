package keys_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"

	"github.com/google/uuid"
	domainapproval "github.com/tokenjoy/backend/internal/domain/approval"
	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func newApprovalEngine(t *testing.T) (*domainapproval.Engine, store.Store) {
	t.Helper()
	svc, st, _ := newKeysServiceWithNewAPI(t)
	repo := st.Approval()
	txRunner := func(ctx context.Context, fn func(store.Store) error) error {
		return st.WithTx(ctx, fn)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	engine := domainapproval.NewEngine(repo, txRunner, logger,
		domainkeys.NewKeyApprovalHandler(svc),
	)
	return engine, st
}

// setReservedPool sets the reserved pool for a department in the test store.
func setReservedPool(t *testing.T, st store.Store, deptID uuid.UUID, reserved int64) {
	t.Helper()
	ctx := testutil.Ctx()
	row, found, err := st.Budget().OrgNodeBudget().Get(ctx, deptID)
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		row = store.OrgNodeBudgetRow{NodeID: deptID, Budget: 99_000_000}
	}
	row.ReservedPool = &reserved
	if err := st.Budget().OrgNodeBudget().Upsert(ctx, deptID, row); err != nil {
		t.Fatal(err)
	}
}

func getReservedPool(t *testing.T, st store.Store, deptID uuid.UUID) int64 {
	t.Helper()
	ctx := testutil.Ctx()
	row, found, err := st.Budget().OrgNodeBudget().Get(ctx, deptID)
	if err != nil {
		t.Fatal(err)
	}
	if !found || row.ReservedPool == nil {
		return 0
	}
	return *row.ReservedPool
}

func TestKeyApproval_DeductsReservedPoolOnTopUp(t *testing.T) {
	t.Parallel()
	engine, st := newApprovalEngine(t)
	ctx := testutil.Ctx()

	// Set a known reserved pool for dept3 (IDMember1's department)
	const initialReserved int64 = 50_000_000_000 // 100k CNY in quota
	setReservedPool(t, st, contract.IDDept3, initialReserved)

	// Zero out member1's personal budget so any key request needs a top-up
	members, err := st.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := range members {
		if members[i].ID == contract.IDMember1 {
			members[i].PersonalBudget = 0
		}
	}
	if err := st.Org().SetMembers(ctx, members); err != nil {
		t.Fatal(err)
	}
	// Remove existing platform keys for member1 so remaining = 0
	allKeys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	filtered := make([]types.PlatformKey, 0, len(allKeys))
	for _, k := range allKeys {
		if k.MemberID != nil && *k.MemberID == contract.IDMember1 {
			continue
		}
		filtered = append(filtered, k)
	}
	if err := st.Keys().SetPlatformKeys(ctx, filtered); err != nil {
		t.Fatal(err)
	}

	// Request a key — since personalBudget=0 and no existing keys, all budget needs top-up
	const requestedBudget int64 = 1_000_000 // 2 CNY
	meta, _ := json.Marshal(types.KeyApprovalMeta{
		Reason:          "need a key",
		RequestedBudget: float64(requestedBudget),
		RequestedModels: []uuid.UUID{contract.IDModel1},
	})
	input := domainapproval.CreateInput{
		Type:           types.ApprovalTypeKey,
		ApplicantID:    contract.IDMember1,
		ApplicantName:  "张三",
		DepartmentID:   contract.IDDept3,
		DepartmentName: "后端组",
		Metadata:       meta,
	}
	req, err := engine.Create(ctx, input)
	if err != nil {
		t.Fatal("create approval:", err)
	}

	// Approve
	err = engine.Approve(ctx, req.ID, domainapproval.ApproverInfo{
		ID:   contract.IDMemberAdmin,
		Name: "Admin",
	})
	if err != nil {
		t.Fatal("approve:", err)
	}

	// Check reserved pool was deducted by the top-up amount
	after := getReservedPool(t, st, contract.IDDept3)
	deducted := initialReserved - after
	if deducted != requestedBudget {
		t.Fatalf("expected reserved pool deduction of %d, got %d (before=%d after=%d)",
			requestedBudget, deducted, initialReserved, after)
	}
}

func TestKeyApproval_NoDeductionWhenSufficientRemaining(t *testing.T) {
	t.Parallel()
	engine, st := newApprovalEngine(t)
	ctx := testutil.Ctx()

	// Give member1 a huge personal budget and remove existing keys
	// so remaining is very large and no top-up is needed
	members, err := st.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := range members {
		if members[i].ID == contract.IDMember1 {
			members[i].PersonalBudget = 999_000_000_000 // ~2M CNY
		}
	}
	if err := st.Org().SetMembers(ctx, members); err != nil {
		t.Fatal(err)
	}
	// Remove existing platform keys for member1 so remaining = personalBudget
	allKeys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
	filtered := make([]types.PlatformKey, 0, len(allKeys))
	for _, k := range allKeys {
		if k.MemberID != nil && *k.MemberID == contract.IDMember1 {
			continue
		}
		filtered = append(filtered, k)
	}
	if err := st.Keys().SetPlatformKeys(ctx, filtered); err != nil {
		t.Fatal(err)
	}

	// Set reserved pool
	const initialReserved int64 = 50_000_000_000
	setReservedPool(t, st, contract.IDDept3, initialReserved)

	// Request a small key that fits within existing budget
	const requestedBudget int64 = 1_000_000
	meta, _ := json.Marshal(types.KeyApprovalMeta{
		Reason:          "small key",
		RequestedBudget: float64(requestedBudget),
		RequestedModels: []uuid.UUID{contract.IDModel1},
	})
	input := domainapproval.CreateInput{
		Type:           types.ApprovalTypeKey,
		ApplicantID:    contract.IDMember1,
		ApplicantName:  "张三",
		DepartmentID:   contract.IDDept3,
		DepartmentName: "后端组",
		Metadata:       meta,
	}
	req, err := engine.Create(ctx, input)
	if err != nil {
		t.Fatal("create:", err)
	}
	err = engine.Approve(ctx, req.ID, domainapproval.ApproverInfo{
		ID:   contract.IDMemberAdmin,
		Name: "Admin",
	})
	if err != nil {
		t.Fatal("approve:", err)
	}

	// Reserved pool should be unchanged
	after := getReservedPool(t, st, contract.IDDept3)
	if after != initialReserved {
		t.Fatalf("expected reserved pool unchanged at %d, got %d", initialReserved, after)
	}
}

func TestKeyApproval_PreApproveRejectsInsufficientPool(t *testing.T) {
	t.Parallel()
	engine, st := newApprovalEngine(t)
	ctx := testutil.Ctx()

	// Set reserved pool very low
	setReservedPool(t, st, contract.IDDept3, 100)

	// Request much more than the pool
	meta, _ := json.Marshal(types.KeyApprovalMeta{
		Reason:          "big key",
		RequestedBudget: float64(999_999_999_999),
		RequestedModels: []uuid.UUID{contract.IDModel1},
	})
	input := domainapproval.CreateInput{
		Type:           types.ApprovalTypeKey,
		ApplicantID:    contract.IDMember1,
		ApplicantName:  "张三",
		DepartmentID:   contract.IDDept3,
		DepartmentName: "后端组",
		Metadata:       meta,
	}
	req, err := engine.Create(ctx, input)
	if err != nil {
		t.Fatal("create:", err)
	}

	// Approve should fail at PreApprove stage
	err = engine.Approve(ctx, req.ID, domainapproval.ApproverInfo{
		ID:   contract.IDMemberAdmin,
		Name: "Admin",
	})
	if err == nil {
		t.Fatal("expected error for insufficient pool, got nil")
	}
}
