package budget

import (
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/exchange"
)

func FindParentNode(nodes []types.BudgetNode, childID string) *types.BudgetNode {
	var parent *types.BudgetNode
	var walk func([]types.BudgetNode) bool
	walk = func(list []types.BudgetNode) bool {
		for i := range list {
			for _, child := range list[i].Children {
				if child.ID == childID {
					parent = &list[i]
					return true
				}
			}
			if len(list[i].Children) > 0 && walk(list[i].Children) {
				return true
			}
		}
		return false
	}
	walk(nodes)
	return parent
}

func ValidateBudgetNodeUpdate(
	tree []types.BudgetNode,
	nodeID string,
	newBudget float64,
	newReservedPool float64,
	projects []types.Project,
	members []types.Member,
) *string {
	node := FindBudgetNode(tree, nodeID)
	if node == nil {
		msg := "Node not found"
		return &msg
	}
	childrenSum := SumChildrenBudget(*node)
	projectSum := ProjectsBudgetForDept(projects, nodeID)
	memberSum := MemberBudgetSumForDept(members, nodeID)
	totalAllocated := childrenSum + newReservedPool + projectSum + memberSum
	if newBudget < totalAllocated {
		msg := fmt.Sprintf("部门预算不能低于已分配总额（子部门¥%.0f + 项目¥%.0f + 成员¥%.0f + 预留池¥%.0f = ¥%.0f）",
			exchange.ToDisplay(childrenSum), exchange.ToDisplay(projectSum), exchange.ToDisplay(memberSum),
			exchange.ToDisplay(newReservedPool), exchange.ToDisplay(totalAllocated))
		return &msg
	}
	parent := FindParentNode(tree, nodeID)
	if parent != nil {
		siblingsSum := 0.0
		for _, child := range parent.Children {
			if child.ID != nodeID {
				siblingsSum += child.Budget
			}
		}
		parentReserved := 0.0
		if parent.ReservedPool != nil {
			parentReserved = *parent.ReservedPool
		}
		if siblingsSum+newReservedPool+newBudget > parent.Budget-parentReserved {
			remaining := parent.Budget - parentReserved - siblingsSum
			if remaining < 0 {
				remaining = 0
			}
			msg := fmt.Sprintf("超出上级可分配预算，当前剩余约 ¥%.0f", exchange.ToDisplay(remaining))
			return &msg
		}
	}
	return nil
}

// ProjectsBudgetForDept returns the sum of project budgets owned by a department.
func ProjectsBudgetForDept(projects []types.Project, deptID string) float64 {
	sum := 0.0
	for _, p := range projects {
		if p.OwnerDepartmentID == deptID {
			sum += p.Budget
		}
	}
	return sum
}

// MemberBudgetSumForDept returns the sum of all members' personal budgets in a department.
func MemberBudgetSumForDept(members []types.Member, deptID string) float64 {
	sum := 0.0
	for _, m := range members {
		if m.DepartmentID == deptID {
			sum += m.PersonalBudget
		}
	}
	return sum
}
