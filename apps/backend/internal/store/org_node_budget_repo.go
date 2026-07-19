package store

import (
	"context"

	"github.com/google/uuid"
)

type OrgNodeBudgetRow struct {
	NodeID          uuid.UUID
	Budget          int64
	ReservedPool    *int64
	Period          string
	MemberAvgBudget int64
}

type OrgNodeBudgetRepository interface {
	Upsert(ctx context.Context, nodeID uuid.UUID, row OrgNodeBudgetRow) error
	UpsertMany(ctx context.Context, rows []OrgNodeBudgetRow) error
	Get(ctx context.Context, nodeID uuid.UUID) (OrgNodeBudgetRow, bool, error)
}
