package company

import (
	"context"

	"github.com/tokenjoy/backend/internal/config"
)

type Gate struct {
	cfg config.Config
}

func NewGate(cfg config.Config) *Gate {
	return &Gate{cfg: cfg}
}

func (g *Gate) IsSuspended(ctx context.Context) bool {
	companyCtx, ok := FromContext(ctx)
	if !ok {
		return false
	}
	return companyCtx.Status == "suspended"
}
