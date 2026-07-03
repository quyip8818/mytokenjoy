package org

import (
	"context"

	"github.com/tokenjoy/backend/internal/store"
)

func (s *service) bumpAuthzRevision(ctx context.Context) error {
	return s.bumpAuthzRevisionStore(ctx, s.store)
}

func (s *service) bumpAuthzRevisionStore(ctx context.Context, st store.Store) error {
	companyID := store.CompanyID(ctx)
	_, err := st.Company().BumpAuthzRevision(ctx, companyID)
	return err
}
