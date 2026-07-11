package worker

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *Runner) processOrgSync(ctx context.Context) error {
	if r.syncSvc == nil || r.companies == nil {
		return nil
	}
	return r.forEachActiveCompany(ctx, func(companyCtx context.Context, _ store.Company) error {
		return r.syncSvc.RunScheduledSync(companyCtx)
	})
}

// forEachActiveCompany runs fn for every active company with a company-scoped
// context. In single-tenant mode this iterates exactly one company, matching
// legacy behavior; in SaaS mode it fans out across all active tenants.
func (r *Runner) forEachActiveCompany(ctx context.Context, fn func(context.Context, store.Company) error) error {
	if r.companies == nil {
		return nil
	}
	companies, err := r.companies.List(ctx)
	if err != nil {
		return err
	}
	for _, co := range companies {
		if co.Status != store.CompanyStatusActive {
			continue
		}
		if err := fn(companyContextFromStore(ctx, co), co); err != nil {
			return err
		}
	}
	return nil
}

func companyContextFromStore(parent context.Context, co store.Company) context.Context {
	info := company.Context{
		CompanyID: co.ID,
		Slug:      co.Slug,
		Status:    co.Status,
	}
	if co.NewAPIWalletUserID != nil {
		info.NewAPIWalletUserID = *co.NewAPIWalletUserID
	}
	return company.WithContext(parent, info)
}
