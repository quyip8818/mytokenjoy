package store

import (
	"context"
	"time"

	"github.com/google/uuid"
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
	TopModelsByDepartments(ctx context.Context, q types.UsageAggregateQuery, deptIDs []uuid.UUID) (map[uuid.UUID]string, error)
}

type NotificationRepository interface {
	Append(ctx context.Context, entry types.NotificationLogEntry) error
	List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]types.NotificationLogEntry, error)
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
	MarkRead(ctx context.Context, id uuid.UUID) error
	MarkAllRead(ctx context.Context, userID uuid.UUID) error
	// Admin queries
	ListLog(ctx context.Context, filter types.NotificationLogFilter) ([]types.NotificationLogEntry, error)
	Stats(ctx context.Context) ([]types.NotificationStatRow, error)
}

type NotificationPreferenceRepository interface {
	Get(ctx context.Context, userID uuid.UUID) ([]types.NotificationPreferenceEntry, error)
	Upsert(ctx context.Context, userID uuid.UUID, entries []types.NotificationPreferenceEntry) error
	Delete(ctx context.Context, userID uuid.UUID) error
}
