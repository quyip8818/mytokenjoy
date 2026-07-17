package budget_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
)

func TestMemberAvgBudgetInheritance(t *testing.T) {
	t.Parallel()

	nodes := []types.OrgNode{
		{
			ID:              "root",
			Name:            "总公司",
			MemberAvgBudget: 500,
			Children: []types.OrgNode{
				{
					ID:              "dept-a",
					Name:            "技术部",
					MemberAvgBudget: 0, // not set — should inherit 500
					Children: []types.OrgNode{
						{
							ID:              "dept-a1",
							Name:            "后端组",
							MemberAvgBudget: 300, // explicitly set
						},
						{
							ID:              "dept-a2",
							Name:            "前端组",
							MemberAvgBudget: 0, // should inherit from dept-a which inherits from root = 500
						},
					},
				},
				{
					ID:              "dept-b",
					Name:            "运营部",
					MemberAvgBudget: 200, // explicitly set
					Children: []types.OrgNode{
						{
							ID:              "dept-b1",
							Name:            "市场组",
							MemberAvgBudget: 0, // should inherit 200 from dept-b
						},
					},
				},
			},
		},
	}

	tree := types.OrgNodesToBudgetTree(nodes)

	// root keeps its own value
	if tree[0].MemberAvgBudget != 500 {
		t.Fatalf("root: expected 500, got %v", tree[0].MemberAvgBudget)
	}

	// dept-a inherits root's 500
	deptA := tree[0].Children[0]
	if deptA.MemberAvgBudget != 500 {
		t.Fatalf("dept-a: expected 500 (inherited), got %v", deptA.MemberAvgBudget)
	}

	// dept-a1 keeps its own 300
	deptA1 := deptA.Children[0]
	if deptA1.MemberAvgBudget != 300 {
		t.Fatalf("dept-a1: expected 300 (own), got %v", deptA1.MemberAvgBudget)
	}

	// dept-a2 inherits dept-a's inherited 500
	deptA2 := deptA.Children[1]
	if deptA2.MemberAvgBudget != 500 {
		t.Fatalf("dept-a2: expected 500 (inherited from dept-a), got %v", deptA2.MemberAvgBudget)
	}

	// dept-b keeps its own 200
	deptB := tree[0].Children[1]
	if deptB.MemberAvgBudget != 200 {
		t.Fatalf("dept-b: expected 200 (own), got %v", deptB.MemberAvgBudget)
	}

	// dept-b1 inherits dept-b's 200
	deptB1 := deptB.Children[0]
	if deptB1.MemberAvgBudget != 200 {
		t.Fatalf("dept-b1: expected 200 (inherited from dept-b), got %v", deptB1.MemberAvgBudget)
	}
}
