package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
)

type LedgerCallFilter struct {
	Model    string
	Status   string
	CallerID string
	Keyword  string
	From     string
	To       string
	Page     int
	PageSize int
}

type LedgerProjectorCursor struct {
	LastOccurredAt *time.Time
	LastLedgerID   *uuid.UUID
	Limit          int
}

type LedgerRepository interface {
	InsertOnConflict(ctx context.Context, entry types.UsageLedgerEntry) (inserted bool, err error)
	InsertSegments(ctx context.Context, entries []types.UsageLedgerEntry) (inserted int, err error)
	ExistsIdempotency(ctx context.Context, idempotencyKey string) (bool, error)
	ListCallSettledPage(ctx context.Context, filter LedgerCallFilter) ([]types.UsageLedgerEntry, int, error)
	CallsSummary(ctx context.Context, filter LedgerCallFilter) (types.CallsSummary, error)
	QueryMinuteSeries(ctx context.Context, q types.UsageSeriesQuery) ([]types.UsageSeriesPoint, error)
	ListCallSettledAfterCursor(ctx context.Context, cursor LedgerProjectorCursor) ([]types.UsageLedgerEntry, error)
	ListCallSettledSince(ctx context.Context, since time.Time) ([]types.UsageLedgerEntry, error)
	SumAmountByDepartment(ctx context.Context, departmentID, periodKey string) (float64, error)
}
