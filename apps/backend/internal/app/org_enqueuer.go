package app

import (
	"context"
	"time"

	domainorgremote "github.com/tokenjoy/backend/internal/domain/org/remote"
	"github.com/tokenjoy/backend/internal/infra/jobs"
)

type orgRiverAdmin interface {
	CancelOrgSyncPending(ctx context.Context, companyID int64) error
}

type orgJobEnqueuer struct {
	enqueuer jobs.Enqueuer
	admin    *OrgRiverAdminHolder
}

func NewOrgEnqueuer(enqueuer jobs.Enqueuer, admin *OrgRiverAdminHolder) domainorgremote.JobEnqueuer {
	if admin == nil {
		admin = NewOrgRiverAdminHolder(nil)
	}
	return orgJobEnqueuer{enqueuer: jobsOrNoop(enqueuer), admin: admin}
}

func (o orgJobEnqueuer) InsertOrgSync(ctx context.Context, companyID int64, scheduledAt *time.Time) error {
	return jobs.InsertOrgSync(ctx, o.enqueuer, nil, companyID, scheduledAt)
}

func (o orgJobEnqueuer) CancelPendingOrgSync(ctx context.Context, companyID int64) error {
	return o.admin.CancelOrgSyncPending(ctx, companyID)
}

var _ domainorgremote.JobEnqueuer = orgJobEnqueuer{}
