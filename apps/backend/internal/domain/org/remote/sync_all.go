package remote

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/store"
)

// RunScheduledSyncAll runs scheduled org sync for every active company.
func (s *Service) RunScheduledSyncAll(ctx context.Context) error {
	return company.ForEachActiveCompany(ctx, s.d.Store.Company(), func(companyCtx context.Context, _ store.Company) error {
		return s.RunScheduledSync(companyCtx)
	})
}
