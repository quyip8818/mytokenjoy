package newapi

import (
	"fmt"
	"strings"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/pkgconst"
)

func HighestModelPriceCNY(models []types.ModelInfo, whitelist []string) float64 {
	allowed := make(map[string]struct{}, len(whitelist))
	for _, name := range whitelist {
		allowed[name] = struct{}{}
	}
	highest := 0.0
	for _, model := range models {
		if len(whitelist) > 0 {
			if _, ok := allowed[model.Name]; !ok {
				continue
			}
		}
		price := model.InputPrice + model.OutputPrice
		if price > highest {
			highest = price
		}
	}
	if highest <= 0 {
		return 1
	}
	return highest
}

func CostCNYFromQuota(quota int64, modelPriceCNY float64) float64 {
	return float64(quota) / float64(pkgconst.QuotaPerUnit) * modelPriceCNY
}

func ToNewAPIUnits(cnyRemaining float64, models []types.ModelInfo, whitelist []string) int64 {
	if cnyRemaining <= 0 {
		return 0
	}
	price := HighestModelPriceCNY(models, whitelist)
	units := cnyRemaining / price * float64(pkgconst.QuotaPerUnit)
	if units < 0 {
		return 0
	}
	return int64(units)
}

func FormatModelLimits(models []string) string {
	if len(models) == 0 {
		return ""
	}
	return strings.Join(models, ",")
}

func ModelPriceCNY(models []types.ModelInfo, modelName string) float64 {
	for _, model := range models {
		if model.Name == modelName {
			price := model.InputPrice + model.OutputPrice
			if price <= 0 {
				return 1
			}
			return price
		}
	}
	return 1
}

func EffectiveWhitelist(keyWhitelist, deptAllowed []string) []string {
	if len(keyWhitelist) == 0 {
		return append([]string{}, deptAllowed...)
	}
	allowed := make(map[string]struct{}, len(deptAllowed))
	for _, name := range deptAllowed {
		allowed[name] = struct{}{}
	}
	out := make([]string, 0, len(keyWhitelist))
	for _, name := range keyWhitelist {
		if _, ok := allowed[name]; ok {
			out = append(out, name)
		}
	}
	return out
}

func RelayGroupForDepartment(departmentID string) string {
	return fmt.Sprintf("%s%s", pkgconst.RelayGroupPrefix, departmentID)
}
