package structure

import "github.com/tokenjoy/backend/internal/domain/org/core"

type Local struct {
	d *core.Deps
}

func New(d *core.Deps) *Local {
	return &Local{d: d}
}
