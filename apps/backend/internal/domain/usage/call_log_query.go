package usage

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type callLogQueryService struct {
	ledger store.LedgerRepository
}

func NewCallLogQuerier(ledger store.LedgerRepository) CallLogQuerier {
	return &callLogQueryService{ledger: ledger}
}

func (s *callLogQueryService) ListCalls(ctx context.Context, params types.AuditCallsQueryParams) (types.PageResult[types.CallLog], error) {
	page, pageSize := types.NormalizePageParams(params.Page, params.PageSize)
	entries, total, err := s.ledger.ListCallSettledPage(ctx, store.LedgerCallFilter{
		Model:    params.Model,
		Status:   params.Status,
		CallerID: params.CallerID,
		Keyword:  params.Keyword,
		From:     params.From,
		To:       params.To,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return types.PageResult[types.CallLog]{}, err
	}
	items := CallLogsFromLedger(entries)
	return types.PageResult[types.CallLog]{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func MinuteSeriesFromLedger(ctx context.Context, ledger store.LedgerRepository, q types.UsageSeriesQuery) (types.UsageSeriesResponse, error) {
	if q.Timezone == "" {
		q.Timezone = ResolveTimezone("")
	}
	if err := ValidateGroupBy(q.GroupBy); err != nil {
		return types.UsageSeriesResponse{}, err
	}
	if err := ValidateWindow(q.Start, q.End, q.Granularity); err != nil {
		return types.UsageSeriesResponse{}, err
	}
	points, err := ledger.QueryMinuteSeries(ctx, q)
	if err != nil {
		return types.UsageSeriesResponse{}, err
	}
	if err := ValidateSeriesPointLimit(len(points)); err != nil {
		return types.UsageSeriesResponse{}, err
	}
	return types.UsageSeriesResponse{
		Granularity: types.UsageGranularityMinute,
		Source:      types.UsageSourceLedger,
		Timezone:    q.Timezone,
		Approximate: false,
		MappingAsOf: types.UsageMappingAsOfIngestTime,
		Points:      points,
	}, nil
}
