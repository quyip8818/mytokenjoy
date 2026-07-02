package budget

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/org"
)

func GetPersonalQuota(members []types.Member, memberID string) float64 {
	member, ok := org.FindMemberByID(members, memberID)
	if !ok {
		return common.DefaultPersonalQuota
	}
	if member.PersonalQuota > 0 {
		return member.PersonalQuota
	}
	return common.DefaultPersonalQuota
}

func AddMemberPersonalQuota(members []types.Member, memberID string, amount float64) []types.Member {
	current := GetPersonalQuota(members, memberID)
	return SetMemberPersonalQuota(members, memberID, current+amount)
}

func SetMemberPersonalQuota(members []types.Member, memberID string, personalQuota float64) []types.Member {
	result := append([]types.Member{}, members...)
	for i := range result {
		if result[i].ID == memberID {
			result[i].PersonalQuota = personalQuota
			return result
		}
	}
	return append(result, types.Member{ID: memberID, PersonalQuota: personalQuota})
}

func GetAllocatedKeyQuota(platformKeys []types.PlatformKey, memberID string) float64 {
	sum := 0.0
	for _, key := range platformKeys {
		if key.MemberID != nil && *key.MemberID == memberID && key.BudgetGroupID == nil && key.Status == "active" {
			sum += key.Quota
		}
	}
	return sum
}

func GetUsedKeyQuota(platformKeys []types.PlatformKey, memberID string) float64 {
	sum := 0.0
	for _, key := range platformKeys {
		if key.MemberID != nil && *key.MemberID == memberID && key.BudgetGroupID == nil && key.Status == "active" {
			sum += key.Used
		}
	}
	return sum
}

func GetQuotaRemaining(members []types.Member, platformKeys []types.PlatformKey, memberID string) float64 {
	remaining := GetPersonalQuota(members, memberID) - GetAllocatedKeyQuota(platformKeys, memberID)
	if remaining < 0 {
		return 0
	}
	return remaining
}

func BuildQuotaSummary(members []types.Member, platformKeys []types.PlatformKey, memberID string, reservedPool float64) types.MemberQuotaSummary {
	totalQuota := GetPersonalQuota(members, memberID)
	used := GetUsedKeyQuota(platformKeys, memberID)
	remaining := GetQuotaRemaining(members, platformKeys, memberID)
	return types.MemberQuotaSummary{
		TotalQuota: totalQuota, Used: used, Remaining: remaining, ReservedPool: reservedPool,
	}
}
