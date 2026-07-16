package adapter

import "github.com/tokenjoy/backend/internal/infra/jobs"

// JobsOrNoop returns the enqueuer if non-nil, otherwise a no-op implementation.
func JobsOrNoop(enqueuer jobs.Enqueuer) jobs.Enqueuer {
	if enqueuer == nil {
		return jobs.NoopEnqueuer{}
	}
	return enqueuer
}
