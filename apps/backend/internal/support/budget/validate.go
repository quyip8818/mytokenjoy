package budget

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
)

func FindParentNode(nodes []types.BudgetNode, childID uuid.UUID) *types.BudgetNode {
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

// ValidationResult holds a structured validation failure from budget checks.
type ValidationResult struct {
	Code    string
	Message string
	Meta    map[string]any
}

func ValidateBudgetNodeUpdate(
	tree []types.BudgetNode,
	nodeID uuid.UUID,
	newBudget float64,
	newReservedPool float64,
	projects []types.Project,
	members []types.Member,
) *ValidationResult {
	node := FindBudgetNode(tree, nodeID)
	if node == nil {
		return &ValidationResult{Message: "Node not found"}
	}
	childrenSum := SumChildrenBudget(*node)
	projectSum := ProjectsBudgetForDept(projects, nodeID)
	memberSum := MemberBudgetSumForDept(members, nodeID)
	totalAllocated := childrenSum + newReservedPool + projectSum + memberSum
	if newBudget < totalAllocated {
		msg := fmt.Sprintf("部门预算不能低于已分配总额（子部门%.2f + 项目%.2f + 成员%.2f + 预留池%.2f = %.2f）",
			childrenSum, projectSum, memberSum, newReservedPool, totalAllocated)
		return &ValidationResult{
			Code:    "BUDGET_BELOW_ALLOCATED",
			Message: msg,
			Meta:    map[string]any{"allocated": totalAllocated, "nodeId": nodeID.String()},
		}
	}
	parent := FindParentNode(tree, nodeID)
	if parent != nil {
		var siblingsSum float64
		for _, child := range parent.Children {
			if child.ID != nodeID {
				siblingsSum += child.Budget
			}
		}
		var parentReserved float64
		if parent.ReservedPool != nil {
			parentReserved = *parent.ReservedPool
		}
		if siblingsSum+newReservedPool+newBudget > parent.Budget-parentReserved {
			remaining := parent.Budget - parentReserved - siblingsSum
			if remaining < 0 {
				remaining = 0
			}
			msg := fmt.Sprintf("超出上级可分配预算，当前剩余约 %g quota", remaining)
			return &ValidationResult{
				Code:    "BUDGET_EXCEED_PARENT",
				Message: msg,
				Meta:    map[string]any{"remaining": remaining, "nodeId": nodeID.String()},
			}
		}
	}
	return nil
}

// ProjectsBudgetForDept returns the sum of project budgets owned by a department.
func ProjectsBudgetForDept(projects []types.Project, deptID uuid.UUID) float64 {
	var sum float64
	for _, p := range projects {
		if p.OwnerDepartmentID == deptID {
			sum += p.Budget
		}
	}
	return sum
}

// MemberBudgetSumForDept returns the sum of all members' personal budgets in a department.
func MemberBudgetSumForDept(members []types.Member, deptID uuid.UUID) float64 {
	var sum float64
	for _, m := range members {
		if m.DepartmentID == deptID {
			sum += m.PersonalBudget
		}
	}
	return sum
}
