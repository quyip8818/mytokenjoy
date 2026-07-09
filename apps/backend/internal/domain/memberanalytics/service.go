package memberanalytics

import (
	"context"

	"github.com/tokenjoy/backend/internal/config"
	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/pkg/clock"
)

type Service interface {
	GetDashboard(ctx context.Context, memberID string) (DashboardView, error)
}

type service struct {
	cfg    config.Config
	keys   domainkeys.Service
	reader domainusage.Reader
	clock  clock.Clock
}

func NewService(
	cfg config.Config,
	keys domainkeys.Service,
	reader domainusage.Reader,
) Service {
	return &service{
		cfg:    cfg,
		keys:   keys,
		reader: reader,
		clock:  cfg.Clock(),
	}
}
