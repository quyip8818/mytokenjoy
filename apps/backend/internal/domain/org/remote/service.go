package remote

import (
	"github.com/tokenjoy/backend/internal/domain/org/core"
)

type Service struct {
	d        *core.Deps
	enqueuer JobEnqueuer
}

func New(d *core.Deps, enqueuer JobEnqueuer) *Service {
	if enqueuer == nil {
		enqueuer = NoopJobEnqueuer
	}
	return &Service{d: d, enqueuer: enqueuer}
}
