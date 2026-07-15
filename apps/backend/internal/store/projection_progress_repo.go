package store

import (
	"context"
	"time"
)

const (
	DashboardProjectionStream = "dashboard_buckets"
)

type ProjectionProgress struct {
	CompanyID      int64
	Stream         string
	LastOccurredAt *time.Time
	LastLedgerID   *string
}

type ProjectionProgressRepository interface {
	Get(ctx context.Context, stream string) (*ProjectionProgress, error)
	GetForUpdate(ctx context.Context, stream string) (*ProjectionProgress, error)
	Advance(ctx context.Context, stream string, lastOccurredAt time.Time, lastLedgerID string) error
}
