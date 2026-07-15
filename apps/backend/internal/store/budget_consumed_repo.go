package store

import "context"

const (
	AxisKindProject     = "project"
	AxisKindPlatformKey = "platform_key"
	AxisKindMember      = "member"
)

// ConsumedDelta represents one axis increment for batch budget_consumed writes.
type ConsumedDelta struct {
	AxisKind  string
	AxisID    string
	PeriodKey string
	Amount    float64
}

type BudgetConsumedRepository interface {
	GetConsumed(ctx context.Context, axisKind, axisID, periodKey string) (float64, bool, error)
	ListConsumed(ctx context.Context, axisKind, periodKey string) (map[string]float64, error)
	ListConsumedByPeriods(ctx context.Context, axisKind string, periodKeys []string) (map[string]map[string]float64, error)
	IncrementConsumed(ctx context.Context, axisKind, axisID, periodKey string, amountPoint float64) error
	// IncrementConsumedBatch atomically increments multiple axes in a single SQL round-trip.
	IncrementConsumedBatch(ctx context.Context, deltas []ConsumedDelta) error
	SetConsumed(ctx context.Context, axisKind, axisID, periodKey string, consumed float64) error
}
