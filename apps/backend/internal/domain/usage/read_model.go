package usage

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/usagequery"
)

type readerService struct {
	callLogQueryService
}

func NewReader(st store.Store) Reader {
	return &readerService{callLogQueryService: callLogQueryService{store: st}}
}

func NewReadModel(st store.Store) ReadModel {
	return NewReader(st)
}

func (s *readerService) UsageMinuteSeries(ctx context.Context, q types.UsageSeriesQuery) (types.UsageSeriesResponse, error) {
	return MinuteSeriesFromLedger(ctx, s.store, q)
}

func (s *readerService) QuerySummary(ctx context.Context, q types.UsageAggregateQuery) (types.UsageSummaryTotals, error) {
	return s.store.Usage().QuerySummary(ctx, q)
}

func (s *readerService) QueryAggregates(ctx context.Context, q types.UsageAggregateQuery) ([]types.UsageAggregateRow, error) {
	return s.store.Usage().QueryAggregates(ctx, q)
}

func (s *readerService) QuerySeries(ctx context.Context, q types.UsageSeriesQuery) ([]types.UsageSeriesPoint, error) {
	return s.store.Usage().QuerySeries(ctx, q)
}

func (s *readerService) TopModelsByDepartments(ctx context.Context, q types.UsageAggregateQuery, deptIDs []string) (map[string]string, error) {
	rows, err := s.store.Usage().QueryFilteredBuckets(ctx, q)
	if err != nil {
		return nil, err
	}
	return usagequery.TopModelPerDepartment(rows, deptIDs), nil
}

var (
	_ ReadModel         = (*readerService)(nil)
	_ AnalyticsQuerier  = (*readerService)(nil)
	_ Reader            = (*readerService)(nil)
)
