package app

import "github.com/tokenjoy/backend/internal/infra/jobs"

func jobsOrNoop(enqueuer jobs.Enqueuer) jobs.Enqueuer {
	if enqueuer == nil {
		return jobs.NoopEnqueuer{}
	}
	return enqueuer
}
