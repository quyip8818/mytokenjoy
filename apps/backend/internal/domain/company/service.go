package company

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/grants"
	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	CreateCompany(ctx context.Context, req CreateCompanyRequest) (CreateCompanyResult, error)
	AcceptInvite(ctx context.Context, req AcceptInviteRequest) (types.Member, error)
	PendingInvitesForUser(ctx context.Context, req PendingInvitesForUserRequest) ([]PendingInvite, error)
	ListCompanies(ctx context.Context) ([]store.Company, error)
	UpdateCompany(ctx context.Context, id uuid.UUID, patch UpdateCompanyPatch) error
	UpgradeToStandard(ctx context.Context, companyID uuid.UUID) error
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
	UserID       uuid.UUID // optional: non-nil → creator becomes super-admin
	Name         string
	Industry     string // optional: 行业
	Size         string // optional: 人员规模
	Type         string // "standard" | "trial" | "selfhosted"
	InviteEmail  string // optional: non-empty → generate invite (platform provisioning)
	MemberAlias  string // optional: alias for the creator member
	MemberAvatar string // optional: avatar for the creator member
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
	Billing() store.BillingRepository
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
	emailSender      EmailSender
}

// EmailSender is the subset of notification.Service that company needs for invite emails.
type EmailSender interface {
	SendDirect(ctx context.Context, channel string, address string, msg domainnotification.RenderedMessage) error
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

// WithEmailSender sets the email sender for invite emails.
func WithEmailSender(sender EmailSender) CompanyServiceOption {
	return func(s *service) {
		if sender != nil {
			s.emailSender = sender
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
	if id == s.cfg.TokenJoyCompanyID || id == s.cfg.LocalCompanyID {
		return domain.Forbidden("protected company cannot be modified")
	}
	if patch.Status != nil {
		if err := s.store.Company().UpdateStatus(ctx, id, *patch.Status); err != nil {
			return err
		}
		s.cacheInvalidator.InvalidateCompany(id)
	}
	return nil
}

// UpgradeToStandard upgrades a trial/demo company to standard.
// Within a transaction: changes type → expires all mock lots → resets wallet.
// After commit: invalidates precheck cache so Gateway rejects test-model immediately.
func (s *service) UpgradeToStandard(ctx context.Context, companyID uuid.UUID) error {
	co, err := s.store.Company().GetByID(ctx, companyID)
	if err != nil {
		return err
	}
	if co == nil {
		return domain.NotFound("company not found")
	}
	if co.Type != store.CompanyTypeTrial && co.Type != store.CompanyTypeDemo {
		return domain.Validation("only trial/demo accounts can be upgraded")
	}

	// Transaction: update type + expire mock lots + reset wallet.
	if err := s.store.WithTx(ctx, func(tx store.Store) error {
		// Lock company row first to serialize with concurrent ingest/consume.
		if _, err := tx.Company().LockForUpdate(ctx, companyID); err != nil {
			return err
		}
		if err := tx.Company().UpdateType(ctx, companyID, store.CompanyTypeStandard); err != nil {
			return err
		}
		if _, err := tx.Billing().ExpireMockLots(ctx, companyID); err != nil {
			return fmt.Errorf("expire mock lots: %w", err)
		}
		remain, err := tx.Billing().SumActiveLotsRemaining(ctx, companyID)
		if err != nil {
			return fmt.Errorf("sum active lots: %w", err)
		}
		return tx.Company().SetWalletQuotaRemain(ctx, companyID, remain, nil)
	}); err != nil {
		return err
	}

	// Post-commit: invalidate precheck cache so test-model is immediately rejected.
	s.cacheInvalidator.InvalidateCompany(companyID)
	return nil
}
