package remote

import (
	"github.com/tokenjoy/backend/internal/domain/org/core"
	"github.com/tokenjoy/backend/internal/infra/jobs"
)

type Service struct {
	d        *core.Deps
	enqueuer jobs.Enqueuer
}

func New(d *core.Deps, enqueuer jobs.Enqueuer) *Service {
	if enqueuer == nil {
		enqueuer = jobs.NoopEnqueuer{}
	}
	return &Service{d: d, enqueuer: enqueuer}
}
