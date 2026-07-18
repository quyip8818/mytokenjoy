package budget_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestListApprovalsSeedsPendingItems(t *testing.T) {
	t.Parallel()
	svc, _ := newBudgetService(t)
	items, err := svc.ListApprovals(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 5 {
		t.Fatalf("expected 5 approvals, got %d", len(items))
	}
	pending := 0
	for _, item := range items {
		if item.Status == "pending" {
			pending++
		}
	}
	if pending != 2 {
		t.Fatalf("expected 2 pending approvals, got %d", pending)
	}
}

func TestResolveApprovalApprove(t *testing.T) {
	t.Parallel()
	svc, _ := newBudgetService(t)
	ctx := testutil.Ctx()
	updated, err := svc.ResolveApproval(ctx, contract.IDBudgetApproval1, types.ResolveBudgetApprovalInput{Status: "approved"})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != "approved" || updated.ResolvedAt == nil {
		t.Fatalf("unexpected updated approval %+v", updated)
	}
	items, err := svc.ListApprovals(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range items {
		if item.ID == contract.IDBudgetApproval1 && item.Status != "approved" {
			t.Fatalf("expected appr-1 approved in list")
		}
	}
}

func TestResolveApprovalRejectRequiresReason(t *testing.T) {
	t.Parallel()
	svc, _ := newBudgetService(t)
	_, err := svc.ResolveApproval(testutil.Ctx(), contract.IDBudgetApproval2, types.ResolveBudgetApprovalInput{Status: "rejected"})
	if err == nil {
		t.Fatal("expected validation error")
	}
}
