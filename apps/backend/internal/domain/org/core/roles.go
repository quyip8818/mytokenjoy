package core

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

func RecalcRoleMemberCounts(ctx context.Context, st store.Store, roles []types.Role) error {
	members, err := st.Org().Members(ctx)
	if err != nil {
		return err
	}
	for i := range roles {
		roles[i].MemberCount = pkgorg.CountMembersByRole(members, roles[i].Name)
	}
	return nil
}
