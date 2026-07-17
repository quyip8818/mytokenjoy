package budget

import (
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
)

func GetReservedPoolForDepartment(tree []types.BudgetNode, departmentID uuid.UUID) float64 {
	node := FindBudgetNode(tree, departmentID)
	if node == nil || node.ReservedPool == nil {
		return 0
	}
	return *node.ReservedPool
}

func GetReservedPoolForMember(tree []types.BudgetNode, members []types.Member, memberID uuid.UUID) float64 {
	for _, member := range members {
		if member.ID == memberID {
			return GetReservedPoolForDepartment(tree, member.DepartmentID)
		}
	}
	return 0
}
