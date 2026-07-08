package org_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
)

func TestRecalcOrgNodeMemberCounts_ExcludesInactive(t *testing.T) {
	nodes := []types.OrgNode{
		{ID: "dept-1", Name: "Engineering", Children: nil},
		{ID: "dept-2", Name: "Sales", Children: nil},
	}

	members := []types.Member{
		{ID: "m1", DepartmentID: "dept-1", Status: types.MemberStatusActive},
		{ID: "m2", DepartmentID: "dept-1", Status: types.MemberStatusInactive},
		{ID: "m3", DepartmentID: "dept-1", Status: types.MemberStatusInactive},
		{ID: "m4", DepartmentID: "dept-2", Status: types.MemberStatusActive},
		{ID: "m5", DepartmentID: "dept-2", Status: types.MemberStatusActive},
		{ID: "m6", DepartmentID: "dept-2", Status: types.MemberStatusPending},
	}

	result := pkgorg.RecalcOrgNodeMemberCounts(nodes, members)

	if len(result) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(result))
	}

	// dept-1: 1 active, 2 inactive → count should be 1
	if result[0].MemberCount != 1 {
		t.Errorf("dept-1: expected MemberCount=1, got %d", result[0].MemberCount)
	}

	// dept-2: 2 active, 1 pending → count should be 3
	if result[1].MemberCount != 3 {
		t.Errorf("dept-2: expected MemberCount=3, got %d", result[1].MemberCount)
	}
}

func TestRecalcOrgNodeMemberCounts_AllInactiveGivesZero(t *testing.T) {
	nodes := []types.OrgNode{
		{ID: "dept-1", Name: "Empty Dept", Children: nil},
	}

	members := []types.Member{
		{ID: "m1", DepartmentID: "dept-1", Status: types.MemberStatusInactive},
		{ID: "m2", DepartmentID: "dept-1", Status: types.MemberStatusInactive},
	}

	result := pkgorg.RecalcOrgNodeMemberCounts(nodes, members)

	if result[0].MemberCount != 0 {
		t.Errorf("expected MemberCount=0 for all-inactive dept, got %d", result[0].MemberCount)
	}
}

func TestRecalcOrgNodeMemberCounts_NestedExcludesInactive(t *testing.T) {
	nodes := []types.OrgNode{
		{
			ID:   "parent",
			Name: "Parent",
			Children: []types.OrgNode{
				{ID: "child", Name: "Child", Children: nil},
			},
		},
	}

	members := []types.Member{
		{ID: "m1", DepartmentID: "child", Status: types.MemberStatusActive},
		{ID: "m2", DepartmentID: "child", Status: types.MemberStatusInactive},
		{ID: "m3", DepartmentID: "parent", Status: types.MemberStatusActive},
	}

	result := pkgorg.RecalcOrgNodeMemberCounts(nodes, members)

	// child: 1 active
	if result[0].Children[0].MemberCount != 1 {
		t.Errorf("child: expected MemberCount=1, got %d", result[0].Children[0].MemberCount)
	}

	// parent: 1 direct active + 1 from child = 2
	if result[0].MemberCount != 2 {
		t.Errorf("parent: expected MemberCount=2, got %d", result[0].MemberCount)
	}
}
