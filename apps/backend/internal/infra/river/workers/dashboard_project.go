package workers

import (
	"context"
	"time"

	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/domain/company"
	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/store"
)

type DashboardProjectWorker struct {
	river.WorkerDefaults[jobs.DashboardProjectArgs]
	projector *domaindashboard.Projector
}

func NewDashboardProjectWorker(projector *domaindashboard.Projector) *DashboardProjectWorker {
	return &DashboardProjectWorker{projector: projector}
}

func (w *DashboardProjectWorker) Work(ctx context.Context, job *river.Job[jobs.DashboardProjectArgs]) error {
	if w.projector == nil {
		return nil
	}
	entryCtx := company.WithDefaultCompany(ctx, job.Args.CompanyID)
	_, err := w.projector.RunBatch(entryCtx, job.Args.CompanyID)
	return err
}

type DashboardReconcileWorker struct {
	river.WorkerDefaults[jobs.DashboardReconcileArgs]
	reconcile *domaindashboard.ReconcileService
	store     store.Store
}

func NewDashboardReconcileWorker(reconcile *domaindashboard.ReconcileService, st store.Store) *DashboardReconcileWorker {
	return &DashboardReconcileWorker{reconcile: reconcile, store: st}
}

func (w *DashboardReconcileWorker) Work(ctx context.Context, job *river.Job[jobs.DashboardReconcileArgs]) error {
	if w.reconcile == nil {
		return nil
	}
	entryCtx := company.WithDefaultCompany(ctx, job.Args.CompanyID)
	if err := w.reconcile.RunCompany(entryCtx, job.Args.CompanyID); err != nil {
		return err
	}
	return w.store.TenantBackgroundState().SetLastDashboardReconcileAt(entryCtx, job.Args.CompanyID, time.Now().UTC())
}
