package app

import (
	"context"

	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/infra/jobs"
)

type budgetJobEnqueuer struct {
	enqueuer jobs.Enqueuer
}

// NewBudgetEnqueuer adapts infra/jobs.Enqueuer to domain/budget.JobEnqueuer.
func NewBudgetEnqueuer(enqueuer jobs.Enqueuer) domainbudget.JobEnqueuer {
	return budgetJobEnqueuer{enqueuer: jobsOrNoop(enqueuer)}
}

func (b budgetJobEnqueuer) InsertBudgetProject(ctx context.Context, companyID int64) error {
	return jobs.InsertBudgetProject(ctx, b.enqueuer, nil, companyID)
}

func (b budgetJobEnqueuer) InsertOverrun(ctx context.Context, companyID int64, payload []byte) error {
	return jobs.InsertOverrun(ctx, b.enqueuer, nil, companyID, payload)
}

func (b budgetJobEnqueuer) InsertRebalance(ctx context.Context, companyID int64, axisKind, axisID string) error {
	return jobs.InsertRebalance(ctx, b.enqueuer, nil, companyID, axisKind, axisID)
}

func (b budgetJobEnqueuer) InsertBudgetReconcile(ctx context.Context, companyID int64) error {
	return jobs.InsertBudgetReconcile(ctx, b.enqueuer, nil, companyID)
}

var _ domainbudget.JobEnqueuer = budgetJobEnqueuer{}
