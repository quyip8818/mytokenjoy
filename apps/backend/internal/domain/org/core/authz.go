package core

import (
	"context"

	"github.com/tokenjoy/backend/internal/store"
)

func BumpAuthzRevision(ctx context.Context, d *Deps) error {
	return BumpAuthzRevisionStore(ctx, d.Store)
}

func BumpAuthzRevisionStore(ctx context.Context, st store.Store) error {
	companyID := store.CompanyID(ctx)
	_, err := st.Company().BumpAuthzRevision(ctx, companyID)
	return err
}
