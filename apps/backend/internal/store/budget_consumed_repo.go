package store

import "context"

const (
	AxisKindProject     = "project"
	AxisKindPlatformKey = "platform_key"
	AxisKindMember      = "member"
)

type BudgetConsumedRepository interface {
	GetConsumed(ctx context.Context, axisKind, axisID, periodKey string) (float64, bool, error)
	ListConsumed(ctx context.Context, axisKind, periodKey string) (map[string]float64, error)
	ListConsumedByPeriods(ctx context.Context, axisKind string, periodKeys []string) (map[string]map[string]float64, error)
	IncrementConsumed(ctx context.Context, axisKind, axisID, periodKey string, amountPoint float64) error
	SetConsumed(ctx context.Context, axisKind, axisID, periodKey string, consumed float64) error
}
