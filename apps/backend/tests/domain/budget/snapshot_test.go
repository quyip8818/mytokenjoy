package budget_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/support/budget"
	"github.com/tokenjoy/backend/internal/support/common"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func currentPeriod() string {
	return pkgbudget.SnapshotKey(pkgbudget.PeriodMonthly, time.Now().UTC())
}

func futurePeriod() string {
	now := time.Now().UTC()
	next := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	return pkgbudget.SnapshotKey(pkgbudget.PeriodMonthly, next)
}

// --- CopyPeriod ---

func TestCopyPeriodSuccess(t *testing.T) {
	t.Parallel()
	svc, st := newBudgetService(t)
	ctx := testutil.Ctx()
	target := futurePeriod()

	if err := svc.CopyPeriod(ctx, target); err != nil {
		t.Fatal(err)
	}
	snap, err := st.BudgetSnapshot().Get(ctx, target)
	if err != nil {
		t.Fatal(err)
	}
	if snap == nil {
		t.Fatal("expected snapshot to be created")
	}
	var payload budget.SnapshotPayload
	if err := json.Unmarshal(snap.Snapshot, &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Tree) == 0 {
		t.Fatal("snapshot tree should not be empty")
	}
	if len(payload.Members) == 0 {
		t.Fatal("snapshot members should not be empty")
	}
}

func TestCopyPeriodRejectsPast(t *testing.T) {
	t.Parallel()
	svc, _ := newBudgetService(t)
	err := svc.CopyPeriod(testutil.Ctx(), "2020-01")
	testutil.AssertDomainStatus(t, err, domain.StatusBadRequest)
}

func TestCopyPeriodRejectsCurrentMonth(t *testing.T) {
	t.Parallel()
	svc, _ := newBudgetService(t)
	err := svc.CopyPeriod(testutil.Ctx(), currentPeriod())
	testutil.AssertDomainStatus(t, err, domain.StatusBadRequest)
}

func TestCopyPeriodRejectsInvalidFormat(t *testing.T) {
	t.Parallel()
	svc, _ := newBudgetService(t)
	err := svc.CopyPeriod(testutil.Ctx(), "not-a-period")
	testutil.AssertDomainStatus(t, err, domain.StatusBadRequest)
}

// --- GetTreeForPeriod ---

func TestGetTreeForPeriodEmpty(t *testing.T) {
	t.Parallel()
	svc, _ := newBudgetService(t)
	tree, err := svc.GetTreeForPeriod(testutil.Ctx(), "2099-12")
	if err != nil {
		t.Fatal(err)
	}
	if tree != nil {
		t.Fatalf("expected nil tree for empty period, got %d nodes", len(tree))
	}
}

func TestGetTreeForPeriodWithSnapshot(t *testing.T) {
	t.Parallel()
	svc, _ := newBudgetService(t)
	ctx := testutil.Ctx()
	target := futurePeriod()

	if err := svc.CopyPeriod(ctx, target); err != nil {
		t.Fatal(err)
	}
	tree, err := svc.GetTreeForPeriod(ctx, target)
	if err != nil {
		t.Fatal(err)
	}
	if len(tree) == 0 {
		t.Fatal("expected non-empty tree from snapshot")
	}
}

// --- UpdateSnapshotNode ---

func TestUpdateSnapshotNodeSuccess(t *testing.T) {
	t.Parallel()
	svc, st := newBudgetService(t)
	ctx := testutil.Ctx()
	target := futurePeriod()

	if err := svc.CopyPeriod(ctx, target); err != nil {
		t.Fatal(err)
	}
	newBudget := 99999.0
	if err := svc.UpdateSnapshotNode(ctx, target, contract.IDDept3, newBudget, nil); err != nil {
		t.Fatal(err)
	}
	// Verify
	snap, _ := st.BudgetSnapshot().Get(ctx, target)
	var payload budget.SnapshotPayload
	if err := json.Unmarshal(snap.Snapshot, &payload); err != nil {
		t.Fatal(err)
	}
	found := findNodeInPayload(payload.Tree, contract.IDDept3)
	if found == nil {
		t.Fatal("dept-3 not found in snapshot after update")
	}
	if found.Budget != newBudget {
		t.Fatalf("expected budget %v, got %v", newBudget, found.Budget)
	}
}

func TestUpdateSnapshotNodeRejectsCurrentPeriod(t *testing.T) {
	t.Parallel()
	svc, _ := newBudgetService(t)
	err := svc.UpdateSnapshotNode(testutil.Ctx(), currentPeriod(), contract.IDDept3, 1000, nil)
	testutil.AssertDomainStatus(t, err, domain.StatusBadRequest)
}

func TestUpdateSnapshotNodeRejectsNoSnapshot(t *testing.T) {
	t.Parallel()
	svc, _ := newBudgetService(t)
	err := svc.UpdateSnapshotNode(testutil.Ctx(), futurePeriod(), contract.IDDept3, 1000, nil)
	testutil.AssertDomainStatus(t, err, domain.StatusNotFound)
}

func TestUpdateSnapshotNodeNotFound(t *testing.T) {
	t.Parallel()
	svc, _ := newBudgetService(t)
	ctx := testutil.Ctx()
	target := futurePeriod()
	if err := svc.CopyPeriod(ctx, target); err != nil {
		t.Fatal(err)
	}
	err := svc.UpdateSnapshotNode(ctx, target, uuid.MustParse("00000000-0000-7000-8000-ffffffffffff"), 1000, nil)
	testutil.AssertDomainStatus(t, err, domain.StatusNotFound)
}

// --- UpdateSnapshotMember ---

func TestUpdateSnapshotMemberSuccess(t *testing.T) {
	t.Parallel()
	svc, st := newBudgetService(t)
	ctx := testutil.Ctx()
	target := futurePeriod()

	if err := svc.CopyPeriod(ctx, target); err != nil {
		t.Fatal(err)
	}
	newBudget := 7777.0
	if err := svc.UpdateSnapshotMember(ctx, target, contract.IDMember1, newBudget); err != nil {
		t.Fatal(err)
	}
	snap, _ := st.BudgetSnapshot().Get(ctx, target)
	var payload budget.SnapshotPayload
	if err := json.Unmarshal(snap.Snapshot, &payload); err != nil {
		t.Fatal(err)
	}
	for _, m := range payload.Members {
		if m.MemberID == contract.IDMember1.String() {
			if m.PersonalBudget != newBudget {
				t.Fatalf("expected member budget %v, got %v", newBudget, m.PersonalBudget)
			}
			return
		}
	}
	t.Fatal("member1 not found in snapshot")
}

func TestUpdateSnapshotMemberNotInSnapshot(t *testing.T) {
	t.Parallel()
	svc, _ := newBudgetService(t)
	ctx := testutil.Ctx()
	target := futurePeriod()
	if err := svc.CopyPeriod(ctx, target); err != nil {
		t.Fatal(err)
	}
	err := svc.UpdateSnapshotMember(ctx, target, uuid.MustParse("00000000-0000-7000-8000-ffffffffffff"), 1000)
	testutil.AssertDomainStatus(t, err, domain.StatusNotFound)
}

// --- UpdateSnapshotProject ---

func TestUpdateSnapshotProjectSuccess(t *testing.T) {
	t.Parallel()
	svc, st := newBudgetService(t)
	ctx := testutil.Ctx()
	target := futurePeriod()

	if err := svc.CopyPeriod(ctx, target); err != nil {
		t.Fatal(err)
	}
	newBudget := 8888.0
	if err := svc.UpdateSnapshotProject(ctx, target, contract.IDProject1, newBudget); err != nil {
		t.Fatal(err)
	}
	snap, _ := st.BudgetSnapshot().Get(ctx, target)
	var payload budget.SnapshotPayload
	if err := json.Unmarshal(snap.Snapshot, &payload); err != nil {
		t.Fatal(err)
	}
	for _, p := range payload.Projects {
		if p.ID == contract.IDProject1 {
			if p.Budget != newBudget {
				t.Fatalf("expected project budget %v, got %v", newBudget, p.Budget)
			}
			return
		}
	}
	t.Fatal("project1 not found in snapshot")
}

// --- No-decrease tests ---

func TestUpdateNodeNoDecreaseCurrentMonth(t *testing.T) {
	t.Parallel()
	svc, st := newBudgetService(t)
	prepareDept3NodeUpdateFixture(t, st)
	ctx := testutil.Ctx()

	// First increase budget to a known value
	reserved := 0.0
	increased := 50000.0
	_, err := svc.UpdateNode(ctx, contract.IDDept3, increased, &reserved)
	if err != nil {
		t.Fatal(err)
	}

	// Now try to decrease — should fail with BUDGET_NO_DECREASE
	_, err = svc.UpdateNode(ctx, contract.IDDept3, increased-1000, &reserved)
	if err == nil {
		t.Fatal("expected error when decreasing current month budget")
	}
	testutil.AssertDomainStatus(t, err, domain.StatusUnprocessable)
}

func TestUpdateMemberBudgetNoDecreaseCurrentMonth(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()
	svc := budget.NewService(cfg, st, common.NewDelayer(false), nil, testutil.TestClock())

	// Get current budget for member1
	budgets, err := svc.ListMemberBudgets(ctx, contract.IDDept3)
	if err != nil {
		t.Fatal(err)
	}
	var currentBudget float64
	for _, b := range budgets {
		if b.MemberID == contract.IDMember1 {
			currentBudget = b.PersonalBudget
			break
		}
	}
	if currentBudget <= 0 {
		t.Skip("member1 has zero budget, cannot test decrease")
	}

	// Try to decrease — should fail
	_, err = svc.UpdateMemberBudget(ctx, contract.IDMember1, currentBudget-1)
	if err == nil {
		t.Fatal("expected error when decreasing current month member budget")
	}
	testutil.AssertDomainStatus(t, err, domain.StatusUnprocessable)
}

// --- helpers ---

func findNodeInPayload(nodes []types.BudgetNode, id uuid.UUID) *types.BudgetNode {
	for i := range nodes {
		if nodes[i].ID == id {
			return &nodes[i]
		}
		if nodes[i].Children != nil {
			if found := findNodeInPayload(nodes[i].Children, id); found != nil {
				return found
			}
		}
	}
	return nil
}
