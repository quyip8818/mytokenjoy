package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	DashboardProjectionStream = "dashboard_buckets"
)

type ProjectionProgress struct {
	CompanyID      uuid.UUID
	Stream         string
	LastOccurredAt *time.Time
	LastLedgerID   *uuid.UUID
}

type ProjectionProgressRepository interface {
	Get(ctx context.Context, stream string) (*ProjectionProgress, error)
	GetForUpdate(ctx context.Context, stream string) (*ProjectionProgress, error)
	Advance(ctx context.Context, stream string, lastOccurredAt time.Time, lastLedgerID uuid.UUID) error
}
