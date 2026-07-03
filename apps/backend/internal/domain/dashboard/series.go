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

	// Read contract: minute granularity reads usage_ledger; hour/day read usage_buckets.
	switch q.Granularity {
	case types.UsageGranularityMinute:
		return s.reader.UsageMinuteSeries(ctx, q)
	case types.UsageGranularityDay, types.UsageGranularityHour:
		return s.seriesFromBuckets(ctx, q, q.Granularity, false)
	default:
		return types.UsageSeriesResponse{}, domainusage.ValidateGroupBy("invalid")
	}
}

func (s *service) seriesFromBuckets(
	ctx context.Context,
	q types.UsageSeriesQuery,
	granularity string,
	approximate bool,
) (types.UsageSeriesResponse, error) {
	bucketQuery := q
	bucketQuery.Granularity = granularity
	points, err := s.reader.QuerySeries(ctx, bucketQuery)
	if err != nil {
		return types.UsageSeriesResponse{}, err
	}
	if err := domainusage.ValidateSeriesPointLimit(len(points)); err != nil {
		return types.UsageSeriesResponse{}, err
	}
	mappingAsOf := types.UsageMappingAsOfIngestTime
	if approximate {
		mappingAsOf = types.UsageMappingAsOfQueryTime
	}
	return types.UsageSeriesResponse{
		Granularity: granularity,
		Source:      types.UsageSourceBuckets,
		Timezone:    q.Timezone,
		Approximate: approximate,
		MappingAsOf: mappingAsOf,
		Points:      points,
	}, nil
}
