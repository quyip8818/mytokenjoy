package org_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
)

var (
	orgDept1  = uuid.MustParse("00000000-0000-7000-0000-000000000d01")
	orgDept2  = uuid.MustParse("00000000-0000-7000-0000-000000000d02")
	orgParent = uuid.MustParse("00000000-0000-7000-0000-000000000d10")
	orgChild  = uuid.MustParse("00000000-0000-7000-0000-000000000d11")

	orgM1 = uuid.MustParse("00000000-0000-7000-0000-000000000e01")
	orgM2 = uuid.MustParse("00000000-0000-7000-0000-000000000e02")
	orgM3 = uuid.MustParse("00000000-0000-7000-0000-000000000e03")
	orgM4 = uuid.MustParse("00000000-0000-7000-0000-000000000e04")
	orgM5 = uuid.MustParse("00000000-0000-7000-0000-000000000e05")
	orgM6 = uuid.MustParse("00000000-0000-7000-0000-000000000e06")
)

func TestRecalcOrgNodeMemberCounts_ExcludesInactive(t *testing.T) {
	nodes := []types.OrgNode{
		{ID: orgDept1, Name: "Engineering", Children: nil},
		{ID: orgDept2, Name: "Sales", Children: nil},
	}

	members := []types.Member{
		{ID: orgM1, DepartmentID: orgDept1, Status: types.MemberStatusActive},
		{ID: orgM2, DepartmentID: orgDept1, Status: types.MemberStatusInactive},
		{ID: orgM3, DepartmentID: orgDept1, Status: types.MemberStatusInactive},
		{ID: orgM4, DepartmentID: orgDept2, Status: types.MemberStatusActive},
		{ID: orgM5, DepartmentID: orgDept2, Status: types.MemberStatusActive},
		{ID: orgM6, DepartmentID: orgDept2, Status: types.MemberStatusPending},
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
		{ID: orgDept1, Name: "Empty Dept", Children: nil},
	}

	members := []types.Member{
		{ID: orgM1, DepartmentID: orgDept1, Status: types.MemberStatusInactive},
		{ID: orgM2, DepartmentID: orgDept1, Status: types.MemberStatusInactive},
	}

	result := pkgorg.RecalcOrgNodeMemberCounts(nodes, members)

	if result[0].MemberCount != 0 {
		t.Errorf("expected MemberCount=0 for all-inactive dept, got %d", result[0].MemberCount)
	}
}

func TestRecalcOrgNodeMemberCounts_NestedExcludesInactive(t *testing.T) {
	nodes := []types.OrgNode{
		{
			ID:   orgParent,
			Name: "Parent",
			Children: []types.OrgNode{
				{ID: orgChild, Name: "Child", Children: nil},
			},
		},
	}

	members := []types.Member{
		{ID: orgM1, DepartmentID: orgChild, Status: types.MemberStatusActive},
		{ID: orgM2, DepartmentID: orgChild, Status: types.MemberStatusInactive},
		{ID: orgM3, DepartmentID: orgParent, Status: types.MemberStatusActive},
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
