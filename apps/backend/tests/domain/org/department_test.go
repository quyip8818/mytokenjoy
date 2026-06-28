package org_test

import (
	"context"
	"errors"
	"testing"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/org"
	"github.com/tokenjoy/backend/internal/pkg/budgetutil"
	"github.com/tokenjoy/backend/internal/pkg/orgutil"
	"github.com/tokenjoy/backend/internal/pkg/routingutil"
	"github.com/tokenjoy/backend/internal/seed"
)

func asDomainError(t *testing.T, err error) *domain.DomainError {
	t.Helper()
	var de *domain.DomainError
	if !errors.As(err, &de) {
		t.Fatalf("expected domain error, got %v", err)
	}
	return de
}

func TestCreateDepartmentPersistsAndProvisions(t *testing.T) {
	svc, st := newTestOrgServiceWithStore(t)

	created, err := svc.CreateDepartment(context.Background(), "Phase0 Team", "dept-2")
	if err != nil {
		t.Fatal(err)
	}
	if created.Name != "Phase0 Team" {
		t.Fatalf("unexpected name %s", created.Name)
	}
	if created.ParentID == nil || *created.ParentID != "dept-2" {
		t.Fatalf("expected parent dept-2, got %v", created.ParentID)
	}

	tree := svc.GetDepartmentTree()
	if orgutil.FindDepartment(tree, created.ID) == nil {
		t.Fatal("created department not found in tree")
	}

	budgetNode := budgetutil.FindBudgetNode(st.Budget().Tree(), created.ID)
	if budgetNode == nil {
		t.Fatal("budget node not created")
	}
	if budgetNode.Budget != 0 {
		t.Fatalf("expected budget 0, got %f", budgetNode.Budget)
	}

	rule := routingutil.GetRoutingRuleForDept(created.ID, st.Models().RoutingRules(), tree)
	if rule == nil {
		t.Fatal("routing rule not created")
	}
	if !rule.Inherited {
		t.Fatal("expected inherited routing rule")
	}
}

func TestUpdateDepartmentPreservesParent(t *testing.T) {
	svc, st := newTestOrgServiceWithStore(t)

	created, err := svc.CreateDepartment(context.Background(), "Rename Me", "dept-6")
	if err != nil {
		t.Fatal(err)
	}

	updated, err := svc.UpdateDepartment(context.Background(), created.ID, "Renamed Team")
	if err != nil {
		t.Fatal(err)
	}
	if updated.ParentID == nil || *updated.ParentID != "dept-6" {
		t.Fatalf("expected parent dept-6, got %v", updated.ParentID)
	}
	if updated.Name != "Renamed Team" {
		t.Fatalf("unexpected name %s", updated.Name)
	}

	budgetNode := budgetutil.FindBudgetNode(st.Budget().Tree(), created.ID)
	if budgetNode == nil || budgetNode.Name != "Renamed Team" {
		t.Fatalf("budget node name not updated: %+v", budgetNode)
	}

	rule := routingutil.GetRoutingRuleForDept(created.ID, st.Models().RoutingRules(), svc.GetDepartmentTree())
	if rule == nil || rule.NodeName != "Renamed Team" {
		t.Fatalf("routing rule name not updated: %+v", rule)
	}
}

func TestDeleteDepartmentWithChildren422(t *testing.T) {
	svc := newTestOrgService(t)
	err := svc.DeleteDepartment(context.Background(), "dept-2")
	de := asDomainError(t, err)
	if de.Status != domain.StatusUnprocessable {
		t.Fatalf("expected 422, got %d", de.Status)
	}
	if de.Message != org.DeptDeleteBlockedMsg {
		t.Fatalf("unexpected message %q", de.Message)
	}
}

func TestDeleteDepartmentWithMembers422(t *testing.T) {
	svc := newTestOrgService(t)
	err := svc.DeleteDepartment(context.Background(), seed.IDDept3)
	de := asDomainError(t, err)
	if de.Status != domain.StatusUnprocessable {
		t.Fatalf("expected 422, got %d", de.Status)
	}
}

func TestDeleteLeafDepartment(t *testing.T) {
	svc, st := newTestOrgServiceWithStore(t)

	created, err := svc.CreateDepartment(context.Background(), "Disposable", "dept-2")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.DeleteDepartment(context.Background(), created.ID); err != nil {
		t.Fatal(err)
	}

	tree := svc.GetDepartmentTree()
	if orgutil.FindDepartment(tree, created.ID) != nil {
		t.Fatal("department still in tree")
	}

	if budgetutil.FindBudgetNode(st.Budget().Tree(), created.ID) != nil {
		t.Fatal("budget node still exists")
	}
	for _, rule := range st.Models().RoutingRules() {
		if rule.NodeID == created.ID {
			t.Fatal("routing rule still exists")
		}
	}
}
