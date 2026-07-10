package usage

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type EntryBuildReader interface {
	Models() store.ModelsRepository
	Audit() store.AuditRepository
	Org() store.OrgRepository
	Keys() store.KeysRepository
}

type Ingestor interface {
	IngestByLogID(ctx context.Context, logID int64, source string) error
}

type Enqueuer interface {
	Enqueue(ctx context.Context, logID int64, source string) error
}

type Queue interface {
	Enqueuer
	RecordFailure(ctx context.Context, logID int64, source string, err error) error
	ApplyRetry(ctx context.Context, job store.IngestJob, ingestErr error) error
}

type CallLogQuerier interface {
	ListCalls(ctx context.Context, params types.AuditCallsQueryParams) (types.PageResult[types.CallLog], error)
}

type ReadModel interface {
	CallLogQuerier
	UsageMinuteSeries(ctx context.Context, q types.UsageSeriesQuery) (types.UsageSeriesResponse, error)
}

type AnalyticsQuerier interface {
	QuerySummary(ctx context.Context, q types.UsageAggregateQuery) (types.UsageSummaryTotals, error)
	QueryAggregates(ctx context.Context, q types.UsageAggregateQuery) ([]types.UsageAggregateRow, error)
	QuerySeries(ctx context.Context, q types.UsageSeriesQuery) ([]types.UsageSeriesPoint, error)
	TopModelsByDepartments(ctx context.Context, q types.UsageAggregateQuery, deptIDs []string) (map[string]string, error)
}

type Reader interface {
	ReadModel
	AnalyticsQuerier
}
