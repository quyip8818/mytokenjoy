package budget

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/org"
)

func GetPersonalBudget(members []types.Member, memberID string) float64 {
	member, ok := org.FindMemberByID(members, memberID)
	if !ok {
		return common.DefaultPersonalBudget
	}
	if member.PersonalBudget > 0 {
		return member.PersonalBudget
	}
	return common.DefaultPersonalBudget
}

func AddMemberPersonalBudget(members []types.Member, memberID string, amount float64) []types.Member {
	current := GetPersonalBudget(members, memberID)
	return SetMemberPersonalBudget(members, memberID, current+amount)
}

func SetMemberPersonalBudget(members []types.Member, memberID string, personalBudget float64) []types.Member {
	result := append([]types.Member{}, members...)
	for i := range result {
		if result[i].ID == memberID {
			result[i].PersonalBudget = personalBudget
			return result
		}
	}
	return append(result, types.Member{ID: memberID, PersonalBudget: personalBudget})
}

func GetAllocatedKeyBudget(platformKeys []types.PlatformKey, memberID string) float64 {
	sum := 0.0
	for _, key := range platformKeys {
		if key.MemberID != nil && *key.MemberID == memberID && key.BudgetGroupID == nil && key.Status == "active" {
			sum += key.Budget
		}
	}
	return sum
}

func GetUsedKeyBudget(platformKeys []types.PlatformKey, memberID string) float64 {
	sum := 0.0
	for _, key := range platformKeys {
		if key.MemberID != nil && *key.MemberID == memberID && key.BudgetGroupID == nil && key.Status == "active" {
			sum += key.Used
		}
	}
	return sum
}

func GetBudgetRemaining(members []types.Member, platformKeys []types.PlatformKey, memberID string) float64 {
	remaining := GetPersonalBudget(members, memberID) - GetAllocatedKeyBudget(platformKeys, memberID)
	if remaining < 0 {
		return 0
	}
	return remaining
}

func BuildBudgetSummary(members []types.Member, platformKeys []types.PlatformKey, memberID string, reservedPool float64) types.MemberBudgetSummary {
	totalQuota := GetPersonalBudget(members, memberID)
	used := GetUsedKeyBudget(platformKeys, memberID)
	remaining := GetBudgetRemaining(members, platformKeys, memberID)
	return types.MemberBudgetSummary{
		TotalBudget: totalQuota, Used: used, Remaining: remaining, ReservedPool: reservedPool,
	}
}
