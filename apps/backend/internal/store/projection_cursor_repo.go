package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	DashboardProjectionStream = "dashboard_buckets"
)

type ProjectionCursor struct {
	CompanyID      uuid.UUID
	Stream         string
	LastOccurredAt *time.Time
	LastLedgerID   *uuid.UUID
}

type ProjectionCursorRepository interface {
	Get(ctx context.Context, stream string) (*ProjectionCursor, error)
	GetForUpdate(ctx context.Context, stream string) (*ProjectionCursor, error)
	Advance(ctx context.Context, stream string, lastOccurredAt time.Time, lastLedgerID uuid.UUID) error
}
