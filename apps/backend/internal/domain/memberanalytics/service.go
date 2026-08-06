package memberanalytics

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/support/clock"
)

type Service interface {
	GetDashboard(ctx context.Context, memberID uuid.UUID) (DashboardView, error)
}

type service struct {
	cfg    config.Config
	budget domainbudget.Service
	reader domainusage.Reader
	clock  clock.Clock
}

func NewService(
	cfg config.Config,
	budget domainbudget.Service,
	reader domainusage.Reader,
	clk clock.Clock,
) Service {
	return &service{
		cfg:    cfg,
		budget: budget,
		reader: reader,
		clock:  clock.OrDefault(clk),
	}
}
