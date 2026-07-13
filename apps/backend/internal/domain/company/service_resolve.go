package company

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *service) ResolveFromMember(ctx context.Context, memberID string) (Context, error) {
	companies, err := s.store.Company().List(ctx)
	if err != nil {
		return Context{}, err
	}
	for _, company := range companies {
		companyCtx := WithContext(ctx, Context{CompanyID: company.ID})
		members, err := s.store.Org().Members(companyCtx)
		if err != nil {
			return Context{}, err
		}
		for _, member := range members {
			if member.ID == memberID {
				return Context{
					CompanyID:          company.ID,
					Slug:               company.Slug,
					NewAPIWalletUserID: newAPIWalletUserIDValue(&company),
					Status:             company.Status,
				}, nil
			}
		}
	}
	return Context{}, domain.NotFound("member not found")
}

func (s *service) ResolveCompanyContext(ctx context.Context, companyID int64) (Context, error) {
	co, err := s.store.Company().GetByID(ctx, companyID)
	if err != nil {
		return Context{}, err
	}
	if co == nil {
		return Context{}, domain.NotFound("company not found")
	}
	return Context{
		CompanyID:          co.ID,
		Slug:               co.Slug,
		NewAPIWalletUserID: newAPIWalletUserIDValue(co),
		Status:             co.Status,
	}, nil
}

func (s *service) ResolveCompanyContextBySlug(ctx context.Context, slug string) (Context, error) {
	if slug == "" {
		return Context{}, domain.BadRequest("company slug required")
	}
	co, err := s.store.Company().GetBySlug(ctx, slug)
	if err != nil {
		return Context{}, err
	}
	if co == nil {
		return Context{}, domain.NotFound("company not found")
	}
	return Context{
		CompanyID:          co.ID,
		Slug:               co.Slug,
		NewAPIWalletUserID: newAPIWalletUserIDValue(co),
		Status:             co.Status,
	}, nil
}

func newAPIWalletUserIDValue(t *store.Company) int64 {
	if t.NewAPIWalletUserID != nil {
		return *t.NewAPIWalletUserID
	}
	return 0
}
