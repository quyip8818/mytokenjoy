package structure

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/grants"
	"github.com/tokenjoy/backend/internal/domain/org/core"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

var protectedRoles = map[string]struct{}{
	grants.RoleSuperAdmin: {},
	grants.RoleOrgAdmin:   {},
}

func validateRolesNotEscalated(roles []string) error {
	for _, role := range roles {
		if _, protected := protectedRoles[role]; protected {
			return domain.Forbidden("cannot assign protected role via member update")
		}
	}
	return nil
}

func mapMemberUniqueError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		if strings.Contains(pgErr.ConstraintName, "email") {
			return domain.Conflict("邮箱已存在")
		}
		if strings.Contains(pgErr.ConstraintName, "phone") {
			return domain.Conflict("手机号已存在")
		}
		return domain.Conflict("成员信息重复")
	}
	return err
}

func persistRecalculatedMemberCounts(ctx context.Context, st store.Store, members []types.Member) error {
	nodes, err := st.Org().Nodes().Tree(ctx)
	if err != nil {
		return err
	}
	nodes = core.RecalcDepartmentMemberCounts(nodes, members)
	return st.Org().Nodes().SetTree(ctx, nodes)
}
