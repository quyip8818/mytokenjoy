package session

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/permission"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	GetByMemberID(ctx context.Context, memberID string) (types.SessionContext, error)
}

type service struct {
	store store.Store
}

func NewService(st store.Store) Service {
	return &service{store: st}
}

func (s *service) GetByMemberID(ctx context.Context, memberID string) (types.SessionContext, error) {
	members, err := s.store.Org().Members(ctx)
	if err != nil {
		return types.SessionContext{}, err
	}
	roles, err := s.store.Org().Roles(ctx)
	if err != nil {
		return types.SessionContext{}, err
	}

	member, ok := pkgorg.FindMemberByID(members, memberID)
	if !ok {
		return types.SessionContext{}, domain.NewDomainError(404, "Member not found")
	}

	permissions := permission.ResolveMemberPermissions(*member, roles)
	return types.SessionContext{
		CompanyID:   member.CompanyID,
		Member:      *member,
		Permissions: permissions,
		ReadOnly:    permission.IsReadOnlySession(permissions),
	}, nil
}
