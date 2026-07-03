package remote

import "github.com/tokenjoy/backend/internal/domain/org/core"

type Service struct {
	d *core.Deps
}

func New(d *core.Deps) *Service {
	return &Service{d: d}
}
