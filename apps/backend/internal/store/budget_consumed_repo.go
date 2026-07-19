package store

import (
	"context"

	"github.com/google/uuid"
)

const (
	AxisKindProject     = "project"
	AxisKindPlatformKey = "platform_key"
	AxisKindMember      = "member"
)

// ConsumedDelta represents one axis increment for batch budget_consumed writes.
type ConsumedDelta struct {
	AxisKind  string
	AxisID    uuid.UUID
	PeriodKey string
	Amount    int64
}

type BudgetConsumedRepository interface {
	GetConsumed(ctx context.Context, axisKind string, axisID uuid.UUID, periodKey string) (int64, bool, error)
	ListConsumed(ctx context.Context, axisKind, periodKey string) (map[uuid.UUID]int64, error)
	ListConsumedByPeriods(ctx context.Context, axisKind string, periodKeys []string) (map[string]map[uuid.UUID]int64, error)
	IncrementConsumed(ctx context.Context, axisKind string, axisID uuid.UUID, periodKey string, amount int64) error
	// IncrementConsumedBatch atomically increments multiple axes in a single SQL round-trip.
	IncrementConsumedBatch(ctx context.Context, deltas []ConsumedDelta) error
	SetConsumed(ctx context.Context, axisKind string, axisID uuid.UUID, periodKey string, consumed int64) error
}
