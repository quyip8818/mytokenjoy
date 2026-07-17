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
	Amount    float64
}

type BudgetConsumedRepository interface {
	GetConsumed(ctx context.Context, axisKind string, axisID uuid.UUID, periodKey string) (float64, bool, error)
	ListConsumed(ctx context.Context, axisKind, periodKey string) (map[uuid.UUID]float64, error)
	ListConsumedByPeriods(ctx context.Context, axisKind string, periodKeys []string) (map[string]map[uuid.UUID]float64, error)
	IncrementConsumed(ctx context.Context, axisKind string, axisID uuid.UUID, periodKey string, amountPoint float64) error
	// IncrementConsumedBatch atomically increments multiple axes in a single SQL round-trip.
	IncrementConsumedBatch(ctx context.Context, deltas []ConsumedDelta) error
	SetConsumed(ctx context.Context, axisKind string, axisID uuid.UUID, periodKey string, consumed float64) error
}
