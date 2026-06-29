package budget

import (
	"github.com/tokenjoy/backend/internal/domain/types"
)

func GetReservedPoolForDepartment(tree []types.BudgetNode, departmentID string) float64 {
	node := FindBudgetNode(tree, departmentID)
	if node == nil || node.ReservedPool == nil {
		return 0
	}
	return *node.ReservedPool
}

func GetReservedPoolForMember(tree []types.BudgetNode, members []types.Member, memberID string) float64 {
	for _, member := range members {
		if member.ID == memberID {
			return GetReservedPoolForDepartment(tree, member.DepartmentID)
		}
	}
	return 0
}
