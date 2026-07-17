package store

import (
	"context"

	"github.com/google/uuid"
)

type OrgNodeBudgetRow struct {
	NodeID          uuid.UUID
	Budget          float64
	ReservedPool    *float64
	Period          string
	MemberAvgBudget float64
}

type OrgNodeBudgetRepository interface {
	Upsert(ctx context.Context, nodeID uuid.UUID, row OrgNodeBudgetRow) error
	UpsertMany(ctx context.Context, rows []OrgNodeBudgetRow) error
	Get(ctx context.Context, nodeID uuid.UUID) (OrgNodeBudgetRow, bool, error)
}
