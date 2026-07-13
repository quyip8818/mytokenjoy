package workers

import (
	"context"
	"time"

	"github.com/riverqueue/river"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/store"
)

type BudgetProjectionWorker struct {
	river.WorkerDefaults[jobs.BudgetProjectionArgs]
	projector *domainbudget.Projector
}

func NewBudgetProjectionWorker(projector *domainbudget.Projector) *BudgetProjectionWorker {
	return &BudgetProjectionWorker{projector: projector}
}

func (w *BudgetProjectionWorker) Work(ctx context.Context, job *river.Job[jobs.BudgetProjectionArgs]) error {
	if w.projector == nil {
		return nil
	}
	entryCtx := company.WithDefaultCompany(ctx, job.Args.CompanyID)
	_, err := w.projector.RunBatch(entryCtx, job.Args.CompanyID)
	return err
}

type BudgetReconcileWorker struct {
	river.WorkerDefaults[jobs.BudgetReconcileArgs]
	reconcile *domainbudget.ReconcileService
	store     store.Store
}

func NewBudgetReconcileWorker(reconcile *domainbudget.ReconcileService, st store.Store) *BudgetReconcileWorker {
	return &BudgetReconcileWorker{reconcile: reconcile, store: st}
}

func (w *BudgetReconcileWorker) Work(ctx context.Context, job *river.Job[jobs.BudgetReconcileArgs]) error {
	if w.reconcile == nil {
		return nil
	}
	entryCtx := company.WithDefaultCompany(ctx, job.Args.CompanyID)
	if err := w.reconcile.RunCompany(entryCtx, job.Args.CompanyID); err != nil {
		return err
	}
	return w.store.TenantBackgroundState().SetLastBudgetReconcileAt(entryCtx, job.Args.CompanyID, time.Now().UTC())
}
