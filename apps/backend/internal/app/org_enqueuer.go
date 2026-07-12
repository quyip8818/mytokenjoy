package app

import (
	"context"

	domainorgremote "github.com/tokenjoy/backend/internal/domain/org/remote"
	"github.com/tokenjoy/backend/internal/infra/jobs"
)

type orgJobEnqueuer struct {
	enqueuer jobs.Enqueuer
}

func NewOrgEnqueuer(enqueuer jobs.Enqueuer) domainorgremote.JobEnqueuer {
	return orgJobEnqueuer{enqueuer: jobsOrNoop(enqueuer)}
}

func (o orgJobEnqueuer) InsertOrgSync(ctx context.Context, companyID int64) error {
	return jobs.InsertOrgSync(ctx, o.enqueuer, nil, companyID)
}

var _ domainorgremote.JobEnqueuer = orgJobEnqueuer{}
