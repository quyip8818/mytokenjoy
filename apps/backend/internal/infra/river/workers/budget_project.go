package workers

import (
	"context"

	"github.com/riverqueue/river"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/infra/jobs"
)

type BudgetProjectWorker struct {
	river.WorkerDefaults[jobs.BudgetProjectArgs]
	projector *domainbudget.Projector
}

func NewBudgetProjectWorker(projector *domainbudget.Projector) *BudgetProjectWorker {
	return &BudgetProjectWorker{projector: projector}
}

func (w *BudgetProjectWorker) Work(ctx context.Context, job *river.Job[jobs.BudgetProjectArgs]) error {
	if w.projector == nil {
		return nil
	}
	_, err := w.projector.RunBatch(ctx, job.Args.CompanyID)
	return err
}

type BudgetReconcileWorker struct {
	river.WorkerDefaults[jobs.BudgetReconcileArgs]
	reconcile *domainbudget.ReconcileService
}

func NewBudgetReconcileWorker(reconcile *domainbudget.ReconcileService) *BudgetReconcileWorker {
	return &BudgetReconcileWorker{reconcile: reconcile}
}

func (w *BudgetReconcileWorker) Work(ctx context.Context, job *river.Job[jobs.BudgetReconcileArgs]) error {
	if w.reconcile == nil {
		return nil
	}
	return w.reconcile.RunCompany(ctx, job.Args.CompanyID)
}

type BudgetReconcileFanoutWorker struct {
	river.WorkerDefaults[jobs.BudgetReconcileFanoutArgs]
	reconcile *domainbudget.ReconcileService
}

func NewBudgetReconcileFanoutWorker(reconcile *domainbudget.ReconcileService) *BudgetReconcileFanoutWorker {
	return &BudgetReconcileFanoutWorker{reconcile: reconcile}
}

func (w *BudgetReconcileFanoutWorker) Work(ctx context.Context, _ *river.Job[jobs.BudgetReconcileFanoutArgs]) error {
	if w.reconcile == nil {
		return nil
	}
	return w.reconcile.FanoutReconcileJobs(ctx)
}
