package org_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestCreateDepartmentPersistsAndProvisions(t *testing.T) {
	t.Parallel()
	svc, st := newTestOrgServiceWithStore(t)
	ctx := testutil.Ctx()

	created, err := svc.CreateDepartment(testutil.Ctx(), "Phase0 Team", "dept-2")
	if err != nil {
		t.Fatal(err)
	}
	if created.Name != "Phase0 Team" {
		t.Fatalf("unexpected name %s", created.Name)
	}
	if created.ParentID == nil || *created.ParentID != "dept-2" {
		t.Fatalf("expected parent dept-2, got %v", created.ParentID)
	}

	tree, err := svc.GetDepartmentTree(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if pkgorg.FindDepartment(tree, created.ID) == nil {
		t.Fatal("created department not found in tree")
	}

	budgetTree, err := common.LoadBudgetTree(ctx, st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	budgetNode := budget.FindBudgetNode(budgetTree, created.ID)
	if budgetNode == nil {
		t.Fatal("budget node not created")
	}
	if budgetNode.Budget != 0 {
		t.Fatalf("expected budget 0, got %f", budgetNode.Budget)
	}

	rules, err := common.LoadRoutingRules(ctx, st.Org().Nodes(), st.Models().Allowlist())
	if err != nil {
		t.Fatal(err)
	}
	rule := common.GetRoutingRuleForDept(created.ID, rules, tree)
	if rule == nil {
		t.Fatal("routing rule not created")
	}
	if !rule.Inherited {
		t.Fatal("expected inherited routing rule")
	}
}

func TestUpdateDepartmentPreservesParent(t *testing.T) {
	t.Parallel()
	svc, st := newTestOrgServiceWithStore(t)
	ctx := testutil.Ctx()

	created, err := svc.CreateDepartment(testutil.Ctx(), "Rename Me", "dept-6")
	if err != nil {
		t.Fatal(err)
	}

	updated, err := svc.UpdateDepartment(testutil.Ctx(), created.ID, "Renamed Team")
	if err != nil {
		t.Fatal(err)
	}
	if updated.ParentID == nil || *updated.ParentID != "dept-6" {
		t.Fatalf("expected parent dept-6, got %v", updated.ParentID)
	}
	if updated.Name != "Renamed Team" {
		t.Fatalf("unexpected name %s", updated.Name)
	}

	budgetTree, err := common.LoadBudgetTree(ctx, st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	budgetNode := budget.FindBudgetNode(budgetTree, created.ID)
	if budgetNode == nil || budgetNode.Name != "Renamed Team" {
		t.Fatalf("budget node name not updated: %+v", budgetNode)
	}

	deptTree, err := svc.GetDepartmentTree(ctx)
	if err != nil {
		t.Fatal(err)
	}
	rules, err := common.LoadRoutingRules(ctx, st.Org().Nodes(), st.Models().Allowlist())
	if err != nil {
		t.Fatal(err)
	}
	rule := common.GetRoutingRuleForDept(created.ID, rules, deptTree)
	if rule == nil || rule.NodeName != "Renamed Team" {
		t.Fatalf("routing rule name not updated: %+v", rule)
	}
}

func TestDeleteDepartmentWithChildren422(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	err := svc.DeleteDepartment(testutil.Ctx(), "dept-2")
	de := asDomainError(t, err)
	if de.Status != domain.StatusUnprocessable {
		t.Fatalf("expected 422, got %d", de.Status)
	}
	if de.Message != domainorg.DeptDeleteBlockedMsg {
		t.Fatalf("unexpected message %q", de.Message)
	}
}

func TestDeleteDepartmentWithMembers422(t *testing.T) {
	t.Parallel()
	svc := newTestOrgService(t)
	err := svc.DeleteDepartment(testutil.Ctx(), contract.IDDept3)
	de := asDomainError(t, err)
	if de.Status != domain.StatusUnprocessable {
		t.Fatalf("expected 422, got %d", de.Status)
	}
}

func TestDeleteLeafDepartment(t *testing.T) {
	t.Parallel()
	svc, st := newTestOrgServiceWithStore(t)
	ctx := testutil.Ctx()

	created, err := svc.CreateDepartment(testutil.Ctx(), "Disposable", "dept-2")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.DeleteDepartment(testutil.Ctx(), created.ID); err != nil {
		t.Fatal(err)
	}

	tree, err := svc.GetDepartmentTree(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if pkgorg.FindDepartment(tree, created.ID) != nil {
		t.Fatal("department still in tree")
	}

	budgetTree, err := common.LoadBudgetTree(ctx, st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	if budget.FindBudgetNode(budgetTree, created.ID) != nil {
		t.Fatal("budget node still exists")
	}
	rules, err := common.LoadRoutingRules(ctx, st.Org().Nodes(), st.Models().Allowlist())
	if err != nil {
		t.Fatal(err)
	}
	for _, rule := range rules {
		if rule.NodeID == created.ID {
			t.Fatal("routing rule still exists")
		}
	}
}
