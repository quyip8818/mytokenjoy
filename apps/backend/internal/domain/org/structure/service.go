package structure

import "github.com/tokenjoy/backend/internal/domain/org/core"

type LocalService struct {
	d *core.Deps
}

func New(d *core.Deps) *LocalService {
	return &LocalService{d: d}
}
