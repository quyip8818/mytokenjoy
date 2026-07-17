package adapter

import (
	"context"
	"time"

	"github.com/google/uuid"
	domainorgremote "github.com/tokenjoy/backend/internal/domain/org/remote"
	"github.com/tokenjoy/backend/internal/infra/jobs"
)

// OrgRiverAdmin is the interface for org-specific River admin operations.
type OrgRiverAdmin interface {
	CancelOrgSyncPending(ctx context.Context, companyID uuid.UUID) error
}

// OrgRiverAdminHolder wraps an OrgRiverAdmin with late-binding support.
type OrgRiverAdminHolder struct {
	inner OrgRiverAdmin
}

// NewOrgRiverAdminHolder creates a new holder. Pass nil for a no-op default.
func NewOrgRiverAdminHolder(initial OrgRiverAdmin) *OrgRiverAdminHolder {
	if initial == nil {
		initial = noopOrgRiverAdmin{}
	}
	return &OrgRiverAdminHolder{inner: initial}
}

// Set replaces the inner admin implementation.
func (h *OrgRiverAdminHolder) Set(admin OrgRiverAdmin) {
	if admin == nil {
		admin = noopOrgRiverAdmin{}
	}
	h.inner = admin
}

// CancelOrgSyncPending delegates to the inner implementation.
func (h *OrgRiverAdminHolder) CancelOrgSyncPending(ctx context.Context, companyID uuid.UUID) error {
	return h.inner.CancelOrgSyncPending(ctx, companyID)
}

type noopOrgRiverAdmin struct{}

func (noopOrgRiverAdmin) CancelOrgSyncPending(context.Context, uuid.UUID) error { return nil }

type orgJobEnqueuer struct {
	enqueuer jobs.Enqueuer
	admin    *OrgRiverAdminHolder
}

// NewOrgEnqueuer adapts infra/jobs.Enqueuer to domain/org/remote.JobEnqueuer.
func NewOrgEnqueuer(enqueuer jobs.Enqueuer, admin *OrgRiverAdminHolder) domainorgremote.JobEnqueuer {
	if admin == nil {
		admin = NewOrgRiverAdminHolder(nil)
	}
	return orgJobEnqueuer{enqueuer: JobsOrNoop(enqueuer), admin: admin}
}

func (o orgJobEnqueuer) InsertOrgSync(ctx context.Context, companyID uuid.UUID, scheduledAt *time.Time) error {
	return jobs.InsertOrgSync(ctx, o.enqueuer, nil, companyID, scheduledAt)
}

func (o orgJobEnqueuer) CancelPendingOrgSync(ctx context.Context, companyID uuid.UUID) error {
	return o.admin.CancelOrgSyncPending(ctx, companyID)
}

var _ domainorgremote.JobEnqueuer = orgJobEnqueuer{}
