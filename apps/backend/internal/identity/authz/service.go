package authz

import (
	"context"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	GetSessionContext(ctx context.Context, companyID int64, memberID string) (types.SessionContext, error)
	RevisionReader
}

type service struct {
	store store.Store
	cache *LRUCache
}

var _ RevisionReader = (*service)(nil)

func NewService(cfg config.Config, st store.Store) Service {
	return &service{
		store: st,
		cache: NewLRUCache(cfg.AuthzCacheSize),
	}
}

func (s *service) GetAuthzRevision(ctx context.Context, companyID int64) (int64, error) {
	return s.store.Company().GetAuthzRevision(ctx, companyID)
}

func (s *service) GetSessionContext(ctx context.Context, companyID int64, memberID string) (types.SessionContext, error) {
	revision, err := s.store.Company().GetAuthzRevision(ctx, companyID)
	if err != nil {
		return types.SessionContext{}, err
	}
	if member, perms, readOnly, ok := s.cache.Get(companyID, memberID, revision); ok {
		return types.SessionContext{
			CompanyID:     companyID,
			AuthzRevision: revision,
			Member:        member,
			Permissions:   perms,
			ReadOnly:      readOnly,
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
		CompanyID:     companyID,
		AuthzRevision: revision,
		Member:        authz.Member,
		Permissions:   permissions,
		ReadOnly:      readOnly,
	}, nil
}
