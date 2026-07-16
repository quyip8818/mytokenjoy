package core

import (
	"context"

	"github.com/tokenjoy/backend/internal/store"
)

func BumpAuthzRevision(ctx context.Context, d *Deps) error {
	return BumpAuthzRevisionStore(ctx, d.Store)
}

// AuthzRevisionBumper is the minimal interface for bumping authz revision.
type AuthzRevisionBumper interface {
	Company() store.CompanyRepository
}

func BumpAuthzRevisionStore(ctx context.Context, st AuthzRevisionBumper) error {
	companyID := store.CompanyID(ctx)
	_, err := st.Company().BumpAuthzRevision(ctx, companyID)
	return err
}
