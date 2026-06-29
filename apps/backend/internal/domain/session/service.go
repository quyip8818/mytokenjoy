package session

import (
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/permission"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	GetByMemberID(memberID string) (types.SessionContext, error)
}

type service struct {
	store store.Store
}

func NewService(st store.Store) Service {
	return &service{store: st}
}

func (s *service) GetByMemberID(memberID string) (types.SessionContext, error) {
	members := s.store.Org().Members()
	roles := s.store.Org().Roles()

	member, ok := pkgorg.FindMemberByID(members, memberID)
	if !ok {
		return types.SessionContext{}, domain.NewDomainError(404, "Member not found")
	}

	permissions := permission.ResolveMemberPermissions(*member, roles)
	return types.SessionContext{
		Member:      *member,
		Permissions: permissions,
		ReadOnly:    permission.IsReadOnlySession(permissions),
	}, nil
}
