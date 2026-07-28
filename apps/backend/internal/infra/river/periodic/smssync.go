package periodic

import (
	"time"

	"github.com/riverqueue/river"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/infra/jobs"
)

// BuildSMSSyncJobs returns periodic jobs for SMS catalog sync.
// Returns nil when River, periodic, or SMS sync is disabled.
func BuildSMSSyncJobs(cfg config.Config) []*river.PeriodicJob {
	if !cfg.RiverEnabled || !cfg.RiverPeriodicEnabled || !cfg.SMSSyncEnabled {
		return nil
	}
	interval := time.Duration(cfg.SMSSyncIntervalSec) * time.Second
	if interval <= 0 {
		interval = 10 * time.Minute
	}
	return []*river.PeriodicJob{
		river.NewPeriodicJob(
			river.PeriodicInterval(interval),
			func() (river.JobArgs, *river.InsertOpts) {
				return jobs.SMSSyncArgs{}, nil
			},
			nil,
		),
	}
}
