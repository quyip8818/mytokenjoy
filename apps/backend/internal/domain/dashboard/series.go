package dashboard

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
)

func (s *service) UsageSeries(ctx context.Context, q types.UsageSeriesQuery, scope domainusage.SessionScope) (types.UsageSeriesResponse, error) {
	if q.Timezone == "" {
		q.Timezone = domainusage.ResolveTimezone("")
	}
	if err := domainusage.ValidateGroupBy(q.GroupBy); err != nil {
		return types.UsageSeriesResponse{}, err
	}
	if err := domainusage.ValidateWindow(q.Start, q.End, q.Granularity); err != nil {
		return types.UsageSeriesResponse{}, err
	}
	scopeDeptIDs, err := s.resolveScope(ctx, scope, q.DepartmentID)
	if err != nil {
		return types.UsageSeriesResponse{}, err
	}
	q.ScopeDeptIDs = scopeDeptIDs

	switch q.Granularity {
	case types.UsageGranularityMinute:
		return s.logAggregator.Series(ctx, q)
	case types.UsageGranularityDay, types.UsageGranularityHour:
		points, err := s.store.Usage().QuerySeries(ctx, q)
		if err != nil {
			return types.UsageSeriesResponse{}, err
		}
		if err := domainusage.ValidateSeriesPointLimit(len(points)); err != nil {
			return types.UsageSeriesResponse{}, err
		}
		return types.UsageSeriesResponse{
			Granularity: q.Granularity,
			Source:      types.UsageSourceBuckets,
			Timezone:    q.Timezone,
			Approximate: false,
			MappingAsOf: types.UsageMappingAsOfIngestTime,
			Points:      points,
		}, nil
	default:
		return types.UsageSeriesResponse{}, domainusage.ValidateGroupBy("invalid")
	}
}
