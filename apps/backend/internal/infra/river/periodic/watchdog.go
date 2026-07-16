package periodic

import (
	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/infra/jobs"
)

func BuildWatchdogJobs(cfg config.Config) []*river.PeriodicJob {
	if !cfg.RiverEnabled || !cfg.RiverPeriodicEnabled {
		return nil
	}
	periodicJobs := []*river.PeriodicJob{
		river.NewPeriodicJob(
			river.PeriodicInterval(cfg.WatchdogInterval()),
			func() (river.JobArgs, *river.InsertOpts) {
				return jobs.TenantWatchdogArgs{}, nil
			},
			nil,
		),
	}
	if cfg.IngestEnabled() {
		periodicJobs = append(periodicJobs, river.NewPeriodicJob(
			river.PeriodicInterval(cfg.IngestReconcileInterval()),
			func() (river.JobArgs, *river.InsertOpts) {
				return jobs.IngestReconcileArgs{}, nil
			},
			nil,
		))
	}
	return periodicJobs
}
