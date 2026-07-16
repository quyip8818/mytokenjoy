package company

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain"
)

func (s *service) ResolveFromMember(ctx context.Context, memberID string) (Context, error) {
	if memberID == "" {
		return Context{}, domain.BadRequest("member id required")
	}
	companyID, err := s.store.Org().FindMemberCompanyID(ctx, memberID)
	if err != nil {
		return Context{}, err
	}
	if companyID == 0 {
		return Context{}, domain.NotFound("member not found")
	}
	return s.ResolveCompanyContext(ctx, companyID)
}

func (s *service) ResolveCompanyContext(ctx context.Context, companyID int64) (Context, error) {
	co, err := s.store.Company().GetByID(ctx, companyID)
	if err != nil {
		return Context{}, err
	}
	if co == nil {
		return Context{}, domain.NotFound("company not found")
	}
	return ContextFromStore(*co), nil
}
