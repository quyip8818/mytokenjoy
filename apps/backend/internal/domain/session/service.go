package session

import (
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/org"
	"github.com/tokenjoy/backend/internal/permission"
	"github.com/tokenjoy/backend/internal/pkg/queryutil"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	GetByMemberID(memberID string) (org.SessionContext, error)
}

type service struct {
	store store.Store
}

func NewService(st store.Store) Service {
	return &service{store: st}
}

func (s *service) GetByMemberID(memberID string) (org.SessionContext, error) {
	members := s.store.Org().Members()
	roles := s.store.Org().Roles()

	member, ok := queryutil.FindMemberByID(members, memberID)
	if !ok {
		return org.SessionContext{}, domain.NewDomainError(404, "Member not found")
	}

	permissions := permission.ResolveMemberPermissions(*member, roles)
	return org.SessionContext{
		Member:      *member,
		Permissions: permissions,
		ReadOnly:    permission.IsReadOnlySession(permissions),
	}, nil
}
