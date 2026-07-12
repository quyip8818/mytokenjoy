//go:build testhook

package middleware_test

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/store"
)

type stubCompanyService struct {
	resolve func(ctx context.Context, companyID int64) (domaincompany.Context, error)
}

func (s *stubCompanyService) ResolveCompanyContext(ctx context.Context, companyID int64) (domaincompany.Context, error) {
	if s.resolve != nil {
		return s.resolve(ctx, companyID)
	}
	return domaincompany.Context{}, domain.NotFound("company not found")
}

type stubCompanyRepo struct {
	revision int64
	err      error
}

func (s *stubCompanyRepo) GetByID(context.Context, int64) (*store.Company, error) {
	panic("stubCompanyRepo: GetByID")
}
func (s *stubCompanyRepo) GetBySlug(context.Context, string) (*store.Company, error) {
	panic("stubCompanyRepo: GetBySlug")
}
func (s *stubCompanyRepo) Create(context.Context, store.Company) error {
	panic("stubCompanyRepo: Create")
}
func (s *stubCompanyRepo) UpdateStatus(context.Context, int64, string) error {
	panic("stubCompanyRepo: UpdateStatus")
}
func (s *stubCompanyRepo) UpdatePackageID(context.Context, int64, *string) error {
	panic("stubCompanyRepo: UpdatePackageID")
}
func (s *stubCompanyRepo) UpdateNewAPIWalletUserID(context.Context, int64, int64) error {
	panic("stubCompanyRepo: UpdateNewAPIWalletUserID")
}
func (s *stubCompanyRepo) UpdateRootDeptID(context.Context, int64, string) error {
	panic("stubCompanyRepo: UpdateRootDeptID")
}
func (s *stubCompanyRepo) GetAuthzRevision(context.Context, int64) (int64, error) {
	return s.revision, s.err
}
func (s *stubCompanyRepo) BumpAuthzRevision(context.Context, int64) (int64, error) {
	panic("stubCompanyRepo: BumpAuthzRevision")
}
func (s *stubCompanyRepo) List(context.Context) ([]store.Company, error) {
	panic("stubCompanyRepo: List")
}
func (s *stubCompanyRepo) LockForUpdate(context.Context, int64) (*store.Company, error) {
	panic("stubCompanyRepo: LockForUpdate")
}
func (s *stubCompanyRepo) ApplyWalletDelta(context.Context, int64, float64, *string) error {
	panic("stubCompanyRepo: ApplyWalletDelta")
}
func (s *stubCompanyRepo) SetWalletRemain(context.Context, int64, float64, *string) error {
	panic("stubCompanyRepo: SetWalletRemain")
}
