package app

import (
	"context"

	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	"github.com/tokenjoy/backend/internal/infra/jobs"
)

type dashboardJobEnqueuer struct {
	enqueuer jobs.Enqueuer
}

func NewDashboardEnqueuer(enqueuer jobs.Enqueuer) domaindashboard.JobEnqueuer {
	return dashboardJobEnqueuer{enqueuer: jobsOrNoop(enqueuer)}
}

func (d dashboardJobEnqueuer) InsertDashboardProject(ctx context.Context, companyID int64) error {
	return jobs.InsertDashboardProject(ctx, d.enqueuer, nil, companyID)
}

func (d dashboardJobEnqueuer) InsertDashboardReconcile(ctx context.Context, companyID int64) error {
	return jobs.InsertDashboardReconcile(ctx, d.enqueuer, nil, companyID)
}

var _ domaindashboard.JobEnqueuer = dashboardJobEnqueuer{}
