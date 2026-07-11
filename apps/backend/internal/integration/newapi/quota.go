package newapi

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/newapiunits"
)

func HighestModelPriceCNY(models []types.ModelInfo, allowedIDs []int64) float64 {
	return newapiunits.HighestModelPriceCNY(models, allowedIDs)
}

func CostFromQuota(quota int64, modelPricePoint float64) float64 {
	return newapiunits.CostFromQuota(quota, modelPricePoint)
}

func ToNewAPIUnits(pointRemaining float64, models []types.ModelInfo, allowedIDs []int64) int64 {
	return newapiunits.ToNewAPIUnits(pointRemaining, models, allowedIDs)
}

func FromNewAPIUnits(units int64, models []types.ModelInfo, allowedIDs []int64) float64 {
	return newapiunits.FromNewAPIUnits(units, models, allowedIDs)
}

func ToQuotaUnits(pointRemaining float64, models []types.ModelInfo, allowedIDs []int64) int64 {
	return newapiunits.ToQuotaUnits(pointRemaining, models, allowedIDs)
}

func FromQuotaUnits(units int64, models []types.ModelInfo, allowedIDs []int64) float64 {
	return newapiunits.FromQuotaUnits(units, models, allowedIDs)
}

func FormatModelLimits(callTypes []string) string {
	return newapiunits.FormatModelLimits(callTypes)
}

func ModelPricePoint(models []types.ModelInfo, allowedIDs []int64, callType string) float64 {
	return newapiunits.ModelPricePoint(models, allowedIDs, callType)
}

func EffectiveWhitelistIDs(keyWhitelist, deptAllowed []int64) []int64 {
	return newapiunits.EffectiveWhitelistIDs(keyWhitelist, deptAllowed)
}

func EffectiveCallTypes(models []types.ModelInfo, allowedIDs []int64) []string {
	return newapiunits.EffectiveCallTypes(models, allowedIDs)
}

func NewAPIGroupForDepartment(departmentID string) string {
	return newapiunits.NewAPIGroupForDepartment(departmentID)
}
