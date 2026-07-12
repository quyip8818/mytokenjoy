package store

import "context"

type OrgNodeBudgetRow struct {
	NodeID          string
	Budget          float64
	ReservedPool    *float64
	Period          string
	MemberAvgBudget float64
}

type OrgNodeBudgetRepository interface {
	Upsert(ctx context.Context, nodeID string, row OrgNodeBudgetRow) error
	UpsertMany(ctx context.Context, rows []OrgNodeBudgetRow) error
	Get(ctx context.Context, nodeID string) (OrgNodeBudgetRow, bool, error)
}
