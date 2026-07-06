package worker

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/company"
)

func (r *Runner) processOrgSync(ctx context.Context) error {
	if r.syncSvc == nil {
		return nil
	}
	return r.syncSvc.RunScheduledSync(company.WithDefaultCompany(ctx, r.cfg.DefaultCompanyID))
}
