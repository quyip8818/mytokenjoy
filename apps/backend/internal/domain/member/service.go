package member

import (
	"context"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
)

type Service interface {
	GetDashboard(ctx context.Context, memberID string) (DashboardView, error)
}

type service struct {
	cfg    config.Config
	keys   domainkeys.Service
	reader domainusage.Reader
	now    func() time.Time
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
		now:    time.Now,
	}
}
