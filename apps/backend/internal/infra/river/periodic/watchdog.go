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
	return []*river.PeriodicJob{
		river.NewPeriodicJob(
			river.PeriodicInterval(cfg.WatchdogInterval()),
			func() (river.JobArgs, *river.InsertOpts) {
				return jobs.TenantWatchdogArgs{}, nil
			},
			nil,
		),
	}
}
