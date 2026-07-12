package store

import (
	"context"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type UsageRepository interface {
	UpsertBucket(ctx context.Context, row types.UsageBucketRow) error
	SetBucket(ctx context.Context, row types.UsageBucketRow) error
	ListBucketsSince(ctx context.Context, since time.Time) ([]types.UsageBucketRow, error)
	QuerySeries(ctx context.Context, q types.UsageSeriesQuery) ([]types.UsageSeriesPoint, error)
	QueryAggregates(ctx context.Context, q types.UsageAggregateQuery) ([]types.UsageAggregateRow, error)
	QuerySummary(ctx context.Context, q types.UsageAggregateQuery) (types.UsageSummaryTotals, error)
	QueryFilteredBuckets(ctx context.Context, q types.UsageAggregateQuery) ([]types.UsageBucketRow, error)
	TopModelsByDepartments(ctx context.Context, q types.UsageAggregateQuery, deptIDs []string) (map[string]string, error)
}

type NotificationRepository interface {
	Append(ctx context.Context, entry types.NotificationLogEntry) error
}
