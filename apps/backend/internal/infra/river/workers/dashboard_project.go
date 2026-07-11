package workers

import (
	"context"

	"github.com/riverqueue/river"
	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	"github.com/tokenjoy/backend/internal/infra/jobs"
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
	_, err := w.projector.RunBatch(ctx, job.Args.CompanyID)
	return err
}

type DashboardProjectFanoutWorker struct {
	river.WorkerDefaults[jobs.DashboardProjectFanoutArgs]
	projector *domaindashboard.Projector
}

func NewDashboardProjectFanoutWorker(projector *domaindashboard.Projector) *DashboardProjectFanoutWorker {
	return &DashboardProjectFanoutWorker{projector: projector}
}

func (w *DashboardProjectFanoutWorker) Work(ctx context.Context, _ *river.Job[jobs.DashboardProjectFanoutArgs]) error {
	if w.projector == nil {
		return nil
	}
	return w.projector.FanoutProjectJobs(ctx)
}

type DashboardReconcileWorker struct {
	river.WorkerDefaults[jobs.DashboardReconcileArgs]
	reconcile *domaindashboard.ReconcileService
}

func NewDashboardReconcileWorker(reconcile *domaindashboard.ReconcileService) *DashboardReconcileWorker {
	return &DashboardReconcileWorker{reconcile: reconcile}
}

func (w *DashboardReconcileWorker) Work(ctx context.Context, job *river.Job[jobs.DashboardReconcileArgs]) error {
	if w.reconcile == nil {
		return nil
	}
	return w.reconcile.RunCompany(ctx, job.Args.CompanyID)
}

type DashboardReconcileFanoutWorker struct {
	river.WorkerDefaults[jobs.DashboardReconcileFanoutArgs]
	reconcile *domaindashboard.ReconcileService
}

func NewDashboardReconcileFanoutWorker(reconcile *domaindashboard.ReconcileService) *DashboardReconcileFanoutWorker {
	return &DashboardReconcileFanoutWorker{reconcile: reconcile}
}

func (w *DashboardReconcileFanoutWorker) Work(ctx context.Context, _ *river.Job[jobs.DashboardReconcileFanoutArgs]) error {
	if w.reconcile == nil {
		return nil
	}
	return w.reconcile.FanoutReconcileJobs(ctx)
}
