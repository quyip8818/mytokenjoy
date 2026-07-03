package usage

import (
	"context"
	"encoding/json"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
)

// Ingestor is the write-path owner for consumption settlement.
// Single entry: ledger insert → projection.Apply → side-effect enqueue (all in one tx).
type Ingestor interface {
	Ingest(ctx context.Context, payload newapi.WebhookLogPayload, source string) error
	IngestFromOutbox(ctx context.Context, raw json.RawMessage) error
	EnqueueFailed(ctx context.Context, payload newapi.WebhookLogPayload, ingestErr error) error
}

// CallLogQuerier is the audit read model for settled calls (usage_ledger only).
type CallLogQuerier interface {
	ListCalls(ctx context.Context, params types.AuditCallsQueryParams) (types.PageResult[types.CallLog], error)
}

// ReadModel is the consumption read surface: audit calls and minute-level series from usage_ledger.
type ReadModel interface {
	CallLogQuerier
	UsageMinuteSeries(ctx context.Context, q types.UsageSeriesQuery) (types.UsageSeriesResponse, error)
}

// AnalyticsQuerier reads usage_buckets aggregates for dashboard analytics.
type AnalyticsQuerier interface {
	QuerySummary(ctx context.Context, q types.UsageAggregateQuery) (types.UsageSummaryTotals, error)
	QueryAggregates(ctx context.Context, q types.UsageAggregateQuery) ([]types.UsageAggregateRow, error)
	QuerySeries(ctx context.Context, q types.UsageSeriesQuery) ([]types.UsageSeriesPoint, error)
	TopModelsByDepartments(ctx context.Context, q types.UsageAggregateQuery, deptIDs []string) (map[string]string, error)
}

// Reader combines ledger read model and bucket analytics for consumption read paths.
type Reader interface {
	ReadModel
	AnalyticsQuerier
}
