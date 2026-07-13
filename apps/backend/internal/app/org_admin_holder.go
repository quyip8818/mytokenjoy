package app

import (
	"context"
)

type OrgRiverAdminHolder struct {
	inner orgRiverAdmin
}

func NewOrgRiverAdminHolder(initial orgRiverAdmin) *OrgRiverAdminHolder {
	if initial == nil {
		initial = noopOrgRiverAdmin{}
	}
	return &OrgRiverAdminHolder{inner: initial}
}

func (h *OrgRiverAdminHolder) Set(admin orgRiverAdmin) {
	if admin == nil {
		admin = noopOrgRiverAdmin{}
	}
	h.inner = admin
}

func (h *OrgRiverAdminHolder) CancelOrgSyncPending(ctx context.Context, companyID int64) error {
	return h.inner.CancelOrgSyncPending(ctx, companyID)
}

type noopOrgRiverAdmin struct{}

func (noopOrgRiverAdmin) CancelOrgSyncPending(context.Context, int64) error { return nil }
