package newapiunits

import (
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
)

func HighestModelPricePoint(models []types.ModelInfo, allowedIDs []uuid.UUID) float64 {
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
		return common.DefaultModelPricePoint
	}
	return highest
}

func CostFromQuota(quota int64, modelPricePoint float64) float64 {
	return float64(quota) / float64(common.QuotaPerUnit) * modelPricePoint
}

func ToNewAPIUnits(pointRemaining float64, models []types.ModelInfo, allowedIDs []uuid.UUID) int64 {
	if pointRemaining <= 0 || math.IsNaN(pointRemaining) || math.IsInf(pointRemaining, 0) {
		return 0
	}
	price := HighestModelPricePoint(models, allowedIDs)
	units := pointRemaining / price * float64(common.QuotaPerUnit)
	if units <= 0 || math.IsNaN(units) {
		return 0
	}
	if math.IsInf(units, 0) || units >= float64(math.MaxInt64) {
		return math.MaxInt64
	}
	return int64(units)
}

func FromNewAPIUnits(units int64, models []types.ModelInfo, allowedIDs []uuid.UUID) float64 {
	if units <= 0 {
		return 0
	}
	price := HighestModelPricePoint(models, allowedIDs)
	return float64(units) / float64(common.QuotaPerUnit) * price
}

func FormatModelLimits(callTypes []string) string {
	if len(callTypes) == 0 {
		return ""
	}
	return strings.Join(callTypes, ",")
}

func ModelPricePoint(models []types.ModelInfo, allowedIDs []uuid.UUID, callType string) float64 {
	if resolved, ok := modelcatalog.ResolveIDForCallType(models, allowedIDs, callType); ok {
		byID := modelcatalog.IndexByID(models)
		if model, found := byID[*resolved]; found {
			price := model.InputPrice + model.OutputPrice
			if price <= 0 {
				return common.DefaultModelPricePoint
			}
			return price
		}
	}
	for _, model := range models {
		if model.Type == callType {
			price := model.InputPrice + model.OutputPrice
			if price <= 0 {
				return common.DefaultModelPricePoint
			}
			return price
		}
	}
	return common.DefaultModelPricePoint
}

func EffectiveWhitelistIDs(keyWhitelist, deptAllowed []uuid.UUID) []uuid.UUID {
	if len(keyWhitelist) == 0 {
		return append([]uuid.UUID{}, deptAllowed...)
	}
	allowed := make(map[uuid.UUID]struct{}, len(deptAllowed))
	for _, id := range deptAllowed {
		allowed[id] = struct{}{}
	}
	out := make([]uuid.UUID, 0, len(keyWhitelist))
	for _, id := range keyWhitelist {
		if _, ok := allowed[id]; ok {
			out = append(out, id)
		}
	}
	return out
}

func NewAPIGroupForDepartment(departmentID uuid.UUID) string {
	return fmt.Sprintf("%s%s", common.NewAPIGroupPrefix, departmentID)
}
