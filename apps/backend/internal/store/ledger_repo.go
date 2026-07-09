package store

import (
	"context"

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

type LedgerRepository interface {
	InsertOnConflict(ctx context.Context, entry types.UsageLedgerEntry) (inserted bool, err error)
	InsertSegments(ctx context.Context, entries []types.UsageLedgerEntry) (inserted int, err error)
	ExistsIdempotency(ctx context.Context, idempotencyKey string) (bool, error)
	ListCallSettledPage(ctx context.Context, filter LedgerCallFilter) ([]types.UsageLedgerEntry, int, error)
	QueryMinuteSeries(ctx context.Context, q types.UsageSeriesQuery) ([]types.UsageSeriesPoint, error)
}
