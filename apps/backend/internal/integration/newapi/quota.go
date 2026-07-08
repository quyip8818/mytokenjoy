package newapi

import (
	"fmt"
	"strings"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
)

func HighestModelPriceCNY(models []types.ModelInfo, allowedIDs []int64) float64 {
	byID := modelcatalog.IndexByID(models)
	highest := 0.0
	for _, id := range allowedIDs {
		model, ok := byID[id]
		if !ok || !model.Enabled {
			continue
		}
		price := model.InputPrice + model.OutputPrice
		if price > highest {
			highest = price
		}
	}
	if len(allowedIDs) == 0 {
		for _, model := range models {
			if !model.Enabled {
				continue
			}
			price := model.InputPrice + model.OutputPrice
			if price > highest {
				highest = price
			}
		}
	}
	if highest <= 0 {
		return common.DefaultModelPriceCNY
	}
	return highest
}

func CostCNYFromQuota(quota int64, modelPriceCNY float64) float64 {
	return float64(quota) / float64(common.QuotaPerUnit) * modelPriceCNY
}

func ToNewAPIUnits(cnyRemaining float64, models []types.ModelInfo, allowedIDs []int64) int64 {
	if cnyRemaining <= 0 {
		return 0
	}
	price := HighestModelPriceCNY(models, allowedIDs)
	units := cnyRemaining / price * float64(common.QuotaPerUnit)
	if units < 0 {
		return 0
	}
	return int64(units)
}

func FromNewAPIUnits(units int64, models []types.ModelInfo, allowedIDs []int64) float64 {
	if units <= 0 {
		return 0
	}
	price := HighestModelPriceCNY(models, allowedIDs)
	return float64(units) / float64(common.QuotaPerUnit) * price
}

func FormatModelLimits(callTypes []string) string {
	if len(callTypes) == 0 {
		return ""
	}
	return strings.Join(callTypes, ",")
}

func ModelPriceCNY(models []types.ModelInfo, allowedIDs []int64, callType string) float64 {
	if resolved, ok := modelcatalog.ResolveIDForCallType(models, allowedIDs, callType); ok {
		byID := modelcatalog.IndexByID(models)
		if model, found := byID[*resolved]; found {
			price := model.InputPrice + model.OutputPrice
			if price <= 0 {
				return common.DefaultModelPriceCNY
			}
			return price
		}
	}
	for _, model := range models {
		if model.Type == callType {
			price := model.InputPrice + model.OutputPrice
			if price <= 0 {
				return common.DefaultModelPriceCNY
			}
			return price
		}
	}
	return common.DefaultModelPriceCNY
}

func EffectiveWhitelistIDs(keyWhitelist, deptAllowed []int64) []int64 {
	if len(keyWhitelist) == 0 {
		return append([]int64{}, deptAllowed...)
	}
	allowed := make(map[int64]struct{}, len(deptAllowed))
	for _, id := range deptAllowed {
		allowed[id] = struct{}{}
	}
	out := make([]int64, 0, len(keyWhitelist))
	for _, id := range keyWhitelist {
		if _, ok := allowed[id]; ok {
			out = append(out, id)
		}
	}
	return out
}

func EffectiveCallTypes(models []types.ModelInfo, allowedIDs []int64) []string {
	return modelcatalog.CallTypesForIDs(models, allowedIDs)
}

func RelayGroupForDepartment(departmentID string) string {
	return fmt.Sprintf("%s%s", common.RelayGroupPrefix, departmentID)
}
