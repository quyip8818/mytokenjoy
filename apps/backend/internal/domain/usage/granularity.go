package usage

import (
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
)

func NormalizeCostGranularity(granularity string) string {
	if granularity == "" {
		return types.UsageGranularityDay
	}
	return granularity
}

func ValidateCostGranularity(granularity string) error {
	if granularity == "" {
		return nil
	}
	switch granularity {
	case types.UsageGranularityDay,
		types.UsageGranularityHour,
		types.UsageGranularityWeek,
		types.UsageGranularityMonth:
		return nil
	default:
		return domain.NewDomainError(domain.StatusBadRequest, "invalid granularity for cost endpoints")
	}
}
