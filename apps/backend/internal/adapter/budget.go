package adapter

import (
	"context"

	"github.com/google/uuid"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/infra/jobs"
)

type budgetJobEnqueuer struct {
	enqueuer jobs.Enqueuer
}

// NewBudgetEnqueuer adapts infra/jobs.Enqueuer to domain/budget.JobEnqueuer.
func NewBudgetEnqueuer(enqueuer jobs.Enqueuer) domainbudget.JobEnqueuer {
	return budgetJobEnqueuer{enqueuer: JobsOrNoop(enqueuer)}
}

func (b budgetJobEnqueuer) InsertOverrun(ctx context.Context, companyID uuid.UUID, payload []byte) error {
	return jobs.InsertOverrun(ctx, b.enqueuer, nil, companyID, payload)
}

func (b budgetJobEnqueuer) InsertRebalance(ctx context.Context, companyID uuid.UUID, axisKind string, axisID uuid.UUID) error {
	return jobs.InsertRebalance(ctx, b.enqueuer, nil, companyID, axisKind, axisID)
}

func (b budgetJobEnqueuer) InsertBudgetReconcile(ctx context.Context, companyID uuid.UUID) error {
	return jobs.InsertBudgetReconcile(ctx, b.enqueuer, nil, companyID)
}

var _ domainbudget.JobEnqueuer = budgetJobEnqueuer{}
