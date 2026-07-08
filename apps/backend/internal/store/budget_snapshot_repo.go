package store

import "context"

const (
	SnapshotAxisOrgNode     = "org_node"
	SnapshotAxisBudgetGroup = "budget_group"
	SnapshotAxisPlatformKey = "platform_key"
	SnapshotAxisMember      = "member"
)

type BudgetSnapshotRepository interface {
	GetConsumed(ctx context.Context, axisKind, axisID, periodKey string) (float64, bool, error)
	ListConsumed(ctx context.Context, axisKind, periodKey string) (map[string]float64, error)
	ListConsumedByPeriods(ctx context.Context, axisKind string, periodKeys []string) (map[string]map[string]float64, error)
	IncrementConsumed(ctx context.Context, axisKind, axisID, periodKey string, amountCNY float64) error
	RollupOrgNodeAncestors(ctx context.Context, leafNodeID, periodKey string, amountCNY float64) error
	SetConsumed(ctx context.Context, axisKind, axisID, periodKey string, consumed float64) error
}
