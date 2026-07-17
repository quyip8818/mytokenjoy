package budget_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
)

func TestMemberAvgBudgetInheritance(t *testing.T) {
	t.Parallel()

	root := uuid.MustParse("00000000-0000-7000-0000-000000000001")
	deptA := uuid.MustParse("00000000-0000-7000-0000-000000000002")
	deptA1 := uuid.MustParse("00000000-0000-7000-0000-000000000003")
	deptA2 := uuid.MustParse("00000000-0000-7000-0000-000000000004")
	deptB := uuid.MustParse("00000000-0000-7000-0000-000000000005")
	deptB1 := uuid.MustParse("00000000-0000-7000-0000-000000000006")

	nodes := []types.OrgNode{
		{
			ID:              root,
			Name:            "总公司",
			MemberAvgBudget: 500,
			Children: []types.OrgNode{
				{
					ID:              deptA,
					Name:            "技术部",
					MemberAvgBudget: 0, // not set — should inherit 500
					Children: []types.OrgNode{
						{
							ID:              deptA1,
							Name:            "后端组",
							MemberAvgBudget: 300, // explicitly set
						},
						{
							ID:              deptA2,
							Name:            "前端组",
							MemberAvgBudget: 0, // should inherit from dept-a which inherits from root = 500
						},
					},
				},
				{
					ID:              deptB,
					Name:            "运营部",
					MemberAvgBudget: 200, // explicitly set
					Children: []types.OrgNode{
						{
							ID:              deptB1,
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
	deptANode := tree[0].Children[0]
	if deptANode.MemberAvgBudget != 500 {
		t.Fatalf("dept-a: expected 500 (inherited), got %v", deptANode.MemberAvgBudget)
	}

	// dept-a1 keeps its own 300
	deptA1Node := deptANode.Children[0]
	if deptA1Node.MemberAvgBudget != 300 {
		t.Fatalf("dept-a1: expected 300 (own), got %v", deptA1Node.MemberAvgBudget)
	}

	// dept-a2 inherits dept-a's inherited 500
	deptA2Node := deptANode.Children[1]
	if deptA2Node.MemberAvgBudget != 500 {
		t.Fatalf("dept-a2: expected 500 (inherited from dept-a), got %v", deptA2Node.MemberAvgBudget)
	}

	// dept-b keeps its own 200
	deptBNode := tree[0].Children[1]
	if deptBNode.MemberAvgBudget != 200 {
		t.Fatalf("dept-b: expected 200 (own), got %v", deptBNode.MemberAvgBudget)
	}

	// dept-b1 inherits dept-b's 200
	deptB1Node := deptBNode.Children[0]
	if deptB1Node.MemberAvgBudget != 200 {
		t.Fatalf("dept-b1: expected 200 (inherited from dept-b), got %v", deptB1Node.MemberAvgBudget)
	}
}
