package company

import (
	"context"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/grants"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	CreateCompany(ctx context.Context, req CreateCompanyRequest) (CreateCompanyResult, error)
	AcceptInvite(ctx context.Context, req AcceptInviteRequest) (types.Member, error)
	ListCompanies(ctx context.Context) ([]store.Company, error)
	UpdateCompany(ctx context.Context, id int64, patch UpdateCompanyPatch) error
	ResolveCompanyContext(ctx context.Context, companyID int64) (Context, error)
	ResolveFromMember(ctx context.Context, memberID string) (Context, error)
}

type UpdateCompanyPatch struct {
	Status    *string
	PackageID *string
}

type CreateCompanyRequest struct {
	Name            string
	SuperAdminEmail string
	PackageID       *string
}

type CreateCompanyResult struct {
	Company    store.Company
	InviteCode string
}

type AcceptInviteRequest struct {
	InviteCode string
	Name       string
	Password   string
}

// Store is the narrow store surface the company domain needs.
type Store interface {
	Company() store.CompanyRepository
	Org() store.OrgRepository
	Invite() store.InviteRepository
	TenantBackgroundState() store.TenantBackgroundStateRepository
	Audit() store.AuditRepository
	WithTx(ctx context.Context, fn func(store.Store) error) error
}

type service struct {
	cfg              config.Config
	store            Store
	client           adminport.Port
	grants           grants.Normalizer
	cacheInvalidator types.PrecheckCacheInvalidator
}

// CompanyServiceOption configures optional dependencies.
type CompanyServiceOption func(*service)

// WithCompanyCacheInvalidator sets the gateway precheck cache invalidator.
func WithCompanyCacheInvalidator(inv types.PrecheckCacheInvalidator) CompanyServiceOption {
	return func(s *service) {
		if inv != nil {
			s.cacheInvalidator = inv
		}
	}
}

func NewService(cfg config.Config, st Store, client adminport.Port, grants grants.Normalizer, opts ...CompanyServiceOption) Service {
	s := &service{cfg: cfg, store: st, client: client, grants: grants, cacheInvalidator: types.NoopPrecheckCacheInvalidator{}}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *service) ListCompanies(ctx context.Context) ([]store.Company, error) {
	return s.store.Company().List(ctx)
}

func (s *service) UpdateCompany(ctx context.Context, id int64, patch UpdateCompanyPatch) error {
	if patch.Status != nil {
		if err := s.store.Company().UpdateStatus(ctx, id, *patch.Status); err != nil {
			return err
		}
		s.cacheInvalidator.InvalidateCompany(id)
	}
	if patch.PackageID != nil {
		if err := s.store.Company().UpdatePackageID(ctx, id, patch.PackageID); err != nil {
			return err
		}
	}
	return nil
}
