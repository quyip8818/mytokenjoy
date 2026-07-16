package adapter

import (
	"context"

	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	"github.com/tokenjoy/backend/internal/infra/jobs"
)

type dashboardJobEnqueuer struct {
	enqueuer jobs.Enqueuer
}

// NewDashboardEnqueuer adapts infra/jobs.Enqueuer to domain/dashboard.JobEnqueuer.
func NewDashboardEnqueuer(enqueuer jobs.Enqueuer) domaindashboard.JobEnqueuer {
	return dashboardJobEnqueuer{enqueuer: JobsOrNoop(enqueuer)}
}

func (d dashboardJobEnqueuer) InsertDashboardProject(ctx context.Context, companyID int64) error {
	return jobs.InsertDashboardProject(ctx, d.enqueuer, nil, companyID)
}

func (d dashboardJobEnqueuer) InsertDashboardReconcile(ctx context.Context, companyID int64) error {
	return jobs.InsertDashboardReconcile(ctx, d.enqueuer, nil, companyID)
}

var _ domaindashboard.JobEnqueuer = dashboardJobEnqueuer{}
