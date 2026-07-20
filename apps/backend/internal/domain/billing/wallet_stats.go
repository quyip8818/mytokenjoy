package billing

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
)

func (s *service) lifetimeRequestCount(ctx context.Context) (int64, error) {
	totals, err := s.reader.QuerySummary(ctx, types.UsageAggregateQuery{
		Start:    types.UsageLifetimeStart,
		End:      types.UsageLifetimeEnd,
		Timezone: types.UsageDefaultTimezone,
	})
	if err != nil {
		return 0, fmt.Errorf("query lifetime wallet stats: %w", err)
	}
	return int64(totals.CallCount), nil
}
