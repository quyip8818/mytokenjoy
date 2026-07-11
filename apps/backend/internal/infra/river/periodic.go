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
				return jobs.OrgSyncArgs{}, nil
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
	}
}
