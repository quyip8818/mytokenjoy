package riverinfra

import (
	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/infra/jobs"
)

func BuildPeriodicJobs(cfg config.Config) []*river.PeriodicJob {
	if !cfg.RiverEnabled {
		return nil
	}
	return []*river.PeriodicJob{
		river.NewPeriodicJob(
			river.PeriodicInterval(cfg.WorkerOrgSyncInterval()),
			func() (river.JobArgs, *river.InsertOpts) {
				return jobs.OrgSyncArgs{CompanyID: jobs.OrgSyncFanoutCompanyID}, jobs.OrgSyncFanoutInsertOpts()
			},
			nil,
		),
		river.NewPeriodicJob(
			river.PeriodicInterval(cfg.WorkerPollInterval()),
			func() (river.JobArgs, *river.InsertOpts) {
				return jobs.MonthlyRebalanceArgs{}, nil
			},
			nil,
		),
		river.NewPeriodicJob(
			river.PeriodicInterval(cfg.WorkerBudgetReconcileInterval()),
			func() (river.JobArgs, *river.InsertOpts) {
				return jobs.BudgetReconcileFanoutArgs{}, nil
			},
			nil,
		),
		river.NewPeriodicJob(
			river.PeriodicInterval(cfg.WorkerDashboardProjectInterval()),
			func() (river.JobArgs, *river.InsertOpts) {
				return jobs.DashboardProjectFanoutArgs{}, nil
			},
			nil,
		),
		river.NewPeriodicJob(
			river.PeriodicInterval(cfg.WorkerDashboardReconcileInterval()),
			func() (river.JobArgs, *river.InsertOpts) {
				return jobs.DashboardReconcileFanoutArgs{}, nil
			},
			nil,
		),
	}
}
