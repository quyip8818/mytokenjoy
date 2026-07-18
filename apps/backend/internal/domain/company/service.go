package company

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/grants"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	CreateCompany(ctx context.Context, req CreateCompanyRequest) (CreateCompanyResult, error)
	AcceptInvite(ctx context.Context, req AcceptInviteRequest) (types.Member, error)
	PendingInvitesForUser(ctx context.Context, req PendingInvitesForUserRequest) ([]PendingInvite, error)
	ListCompanies(ctx context.Context) ([]store.Company, error)
	UpdateCompany(ctx context.Context, id uuid.UUID, patch UpdateCompanyPatch) error
	ResolveCompanyContext(ctx context.Context, companyID uuid.UUID) (Context, error)
	ResolveFromMember(ctx context.Context, memberID uuid.UUID) (Context, error)
}

type UpdateCompanyPatch struct {
	Status *string
}

// CreateCompanyRequest supports two modes:
// - UserID mode (UserID != Nil): creator becomes super-admin Member immediately
// - InviteEmail mode (UserID == Nil && InviteEmail != ""): generates invite for deferred join
type CreateCompanyRequest struct {
	UserID      uuid.UUID // optional: non-nil → creator becomes super-admin
	Name        string
	Type        string // "standard" | "trial" | "selfhosted"
	InviteEmail string // optional: non-empty → generate invite (platform provisioning)
}

type CreateCompanyResult struct {
	Company    store.Company
	Member     *types.Member // non-nil in UserID mode
	InviteCode string        // non-empty in InviteEmail mode
}

// AcceptInviteRequest — user identity is resolved by the handler layer.
type AcceptInviteRequest struct {
	UserID     uuid.UUID
	InviteCode string
	Name       string
}

type PendingInvitesForUserRequest struct {
	Email  string
	Phone  string
	UserID uuid.UUID
}

type PendingInvite struct {
	InviteCode  string    `json:"inviteCode"`
	CompanyID   uuid.UUID `json:"companyId"`
	CompanyName string    `json:"companyName"`
	Role        string    `json:"role"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

// Store is the narrow store surface the company domain needs.
type Store interface {
	Company() store.CompanyRepository
	User() store.UserRepository
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

func (s *service) UpdateCompany(ctx context.Context, id uuid.UUID, patch UpdateCompanyPatch) error {
	if patch.Status != nil {
		if err := s.store.Company().UpdateStatus(ctx, id, *patch.Status); err != nil {
			return err
		}
		s.cacheInvalidator.InvalidateCompany(id)
	}
	return nil
}
