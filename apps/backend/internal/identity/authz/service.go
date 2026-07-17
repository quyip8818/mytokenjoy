package authz

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/billing"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/ctxcompany"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	GetSessionContext(ctx context.Context, companyID uuid.UUID, memberID uuid.UUID) (types.SessionContext, error)
	RevisionReader
}

// Store is the narrow store surface the authz service needs.
type Store interface {
	Company() store.CompanyRepository
	Org() store.OrgRepository
	Billing() store.BillingRepository
}

type service struct {
	store    Store
	cache    *LRUCache
	revCache *revisionCache
}

var _ RevisionReader = (*service)(nil)

func NewService(cfg config.Config, st Store) Service {
	return &service{
		store:    st,
		cache:    NewLRUCache(cfg.AuthzCacheSize),
		revCache: newRevisionCache(5 * time.Second),
	}
}

func (s *service) GetAuthzRevision(ctx context.Context, companyID uuid.UUID) (int64, error) {
	return s.store.Company().GetAuthzRevision(ctx, companyID)
}

func (s *service) GetSessionContext(ctx context.Context, companyID uuid.UUID, memberID uuid.UUID) (types.SessionContext, error) {
	revision, err := s.revCache.get(ctx, companyID, s.store.Company())
	if err != nil {
		return types.SessionContext{}, err
	}

	currency, ppu, err := billing.ResolveCompanyChargeRate(ctx, s.store, companyID)
	if err != nil {
		return types.SessionContext{}, err
	}

	// Read company type from request context (injected by CompanyResolve middleware).
	companyType := companyTypeFromContext(ctx, companyID, s.store)

	if member, perms, readOnly, ok := s.cache.Get(companyID, memberID, revision); ok {
		return types.SessionContext{
			CompanyID:       companyID,
			CompanyType:     companyType,
			AuthzRevision:   revision,
			Member:          member,
			Permissions:     perms,
			ReadOnly:        readOnly,
			BillingCurrency: currency,
			PointsPerUnit:   ppu,
		}, nil
	}

	authz, err := s.store.Org().GetMemberAuthz(ctx, companyID, memberID)
	if err != nil {
		return types.SessionContext{}, err
	}
	if authz == nil || authz.Member.Status != types.MemberStatusActive {
		return types.SessionContext{}, domain.NewDomainError(404, "Member not found")
	}
	permissions := ResolveMemberPermissions(authz.Member, authz.Roles)
	readOnly := IsReadOnlySession(permissions)
	s.cache.Put(companyID, memberID, revision, authz.Member, permissions, readOnly)
	return types.SessionContext{
		CompanyID:       companyID,
		CompanyType:     companyType,
		AuthzRevision:   revision,
		Member:          authz.Member,
		Permissions:     permissions,
		ReadOnly:        readOnly,
		BillingCurrency: currency,
		PointsPerUnit:   ppu,
	}, nil
}

// companyTypeFromContext tries to get company type from the request context first
// (already resolved by CompanyResolve middleware), falling back to a DB lookup only if needed.
func companyTypeFromContext(ctx context.Context, companyID uuid.UUID, st Store) string {
	if info, ok := ctxcompany.From(ctx); ok && info.CompanyID == companyID && info.Type != "" {
		return info.Type
	}
	// Fallback: context not available (e.g. tests).
	co, err := st.Company().GetByID(ctx, companyID)
	if err != nil || co == nil {
		return ""
	}
	return co.Type
}

// revisionCache caches authz_revision per company with a short TTL.
// Reduces per-request DB round trips from 1 to ~0 (amortized over TTL window).
type revisionCache struct {
	mu      sync.Mutex
	ttl     time.Duration
	entries map[uuid.UUID]revisionEntry
}

type revisionEntry struct {
	revision  int64
	expiresAt time.Time
}

func newRevisionCache(ttl time.Duration) *revisionCache {
	return &revisionCache{
		ttl:     ttl,
		entries: make(map[uuid.UUID]revisionEntry),
	}
}

func (rc *revisionCache) get(ctx context.Context, companyID uuid.UUID, repo store.CompanyRepository) (int64, error) {
	now := time.Now()
	rc.mu.Lock()
	if entry, ok := rc.entries[companyID]; ok && now.Before(entry.expiresAt) {
		rc.mu.Unlock()
		return entry.revision, nil
	}
	rc.mu.Unlock()

	revision, err := repo.GetAuthzRevision(ctx, companyID)
	if err != nil {
		return 0, err
	}

	rc.mu.Lock()
	rc.entries[companyID] = revisionEntry{revision: revision, expiresAt: now.Add(rc.ttl)}
	rc.mu.Unlock()
	return revision, nil
}
