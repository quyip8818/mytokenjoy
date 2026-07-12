package usage

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type readerService struct {
	callLogQueryService
	usage store.UsageRepository
}

func NewReader(usage store.UsageRepository, ledger store.LedgerRepository) Reader {
	return &readerService{
		callLogQueryService: callLogQueryService{ledger: ledger},
		usage:               usage,
	}
}

func NewReadModel(usage store.UsageRepository, ledger store.LedgerRepository) ReadModel {
	return NewReader(usage, ledger)
}

func (s *readerService) UsageMinuteSeries(ctx context.Context, q types.UsageSeriesQuery) (types.UsageSeriesResponse, error) {
	return MinuteSeriesFromLedger(ctx, s.ledger, q)
}

func (s *readerService) QuerySummary(ctx context.Context, q types.UsageAggregateQuery) (types.UsageSummaryTotals, error) {
	return s.usage.QuerySummary(ctx, q)
}

func (s *readerService) QueryAggregates(ctx context.Context, q types.UsageAggregateQuery) ([]types.UsageAggregateRow, error) {
	return s.usage.QueryAggregates(ctx, q)
}

func (s *readerService) QuerySeries(ctx context.Context, q types.UsageSeriesQuery) ([]types.UsageSeriesPoint, error) {
	return s.usage.QuerySeries(ctx, q)
}

func (s *readerService) TopModelsByDepartments(ctx context.Context, q types.UsageAggregateQuery, deptIDs []string) (map[string]string, error) {
	return s.usage.TopModelsByDepartments(ctx, q, deptIDs)
}

var (
	_ ReadModel        = (*readerService)(nil)
	_ AnalyticsQuerier = (*readerService)(nil)
	_ Reader           = (*readerService)(nil)
)
