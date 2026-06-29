package budget

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

func GetPersonalQuota(pools map[string]types.MemberQuotaPool, memberID string) float64 {
	if pool, ok := pools[memberID]; ok {
		return pool.PersonalQuota
	}
	return common.DefaultPersonalQuota
}

func AddPersonalQuota(pools map[string]types.MemberQuotaPool, memberID string, amount float64) {
	current := GetPersonalQuota(pools, memberID)
	pools[memberID] = types.MemberQuotaPool{PersonalQuota: current + amount}
}

func SetPersonalQuota(pools map[string]types.MemberQuotaPool, memberID string, personalQuota float64) {
	pools[memberID] = types.MemberQuotaPool{PersonalQuota: personalQuota}
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

func GetQuotaRemaining(pools map[string]types.MemberQuotaPool, platformKeys []types.PlatformKey, memberID string) float64 {
	remaining := GetPersonalQuota(pools, memberID) - GetAllocatedKeyQuota(platformKeys, memberID)
	if remaining < 0 {
		return 0
	}
	return remaining
}

func BuildQuotaSummary(pools map[string]types.MemberQuotaPool, platformKeys []types.PlatformKey, memberID string, reservedPool float64) types.MemberQuotaSummary {
	totalQuota := GetPersonalQuota(pools, memberID)
	used := GetUsedKeyQuota(platformKeys, memberID)
	remaining := GetQuotaRemaining(pools, platformKeys, memberID)
	return types.MemberQuotaSummary{
		TotalQuota: totalQuota, Used: used, Remaining: remaining, ReservedPool: reservedPool,
	}
}
