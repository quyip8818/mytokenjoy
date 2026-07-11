package budget

import (
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
)

func GetAllocatedGroupKeyBudget(platformKeys []types.PlatformKey, budgetGroupID string) float64 {
	sum := 0.0
	for _, key := range platformKeys {
		if key.BudgetGroupID != nil && *key.BudgetGroupID == budgetGroupID && key.Status == "active" {
			sum += key.Budget
		}
	}
	return sum
}

func GetGroupBudgetRemaining(group types.BudgetGroup, platformKeys []types.PlatformKey) float64 {
	allocated := GetAllocatedGroupKeyBudget(platformKeys, group.ID)
	remaining := group.Budget - group.Consumed - allocated
	if remaining < 0 {
		return 0
	}
	return remaining
}

func ValidateGroupKeyBudget(group types.BudgetGroup, platformKeys []types.PlatformKey, budget float64, excludeKeyID string) *string {
	allocated := 0.0
	for _, key := range platformKeys {
		if key.BudgetGroupID != nil && *key.BudgetGroupID == group.ID && key.Status == "active" {
			if excludeKeyID != "" && key.ID == excludeKeyID {
				continue
			}
			allocated += key.Budget
		}
	}
	remaining := group.Budget - group.Consumed - allocated
	if budget > remaining {
		display := remaining
		if display < 0 {
			display = 0
		}
		msg := fmt.Sprintf("预算组剩余可分配额度约 ¥%.0f", display)
		return &msg
	}
	return nil
}
