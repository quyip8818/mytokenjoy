package structure

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/grants"
	"github.com/tokenjoy/backend/internal/domain/org/core"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/ctxcompany"
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

// orgStore is the minimal interface for org node persistence.
type orgStore interface {
	Org() store.OrgRepository
}

func persistRecalculatedMemberCounts(ctx context.Context, st orgStore, members []types.Member) error {
	nodes, err := st.Org().Nodes().Tree(ctx)
	if err != nil {
		return err
	}
	nodes = core.RecalcDepartmentMemberCounts(nodes, members)
	return st.Org().Nodes().SetTree(ctx, nodes)
}

// checkTrialMemberLimit enforces the trial member limit.
// Only applies when the company is type "trial" and a limit is configured.
func (s *LocalService) checkTrialMemberLimit(ctx context.Context, members []types.Member) error {
	return s.checkTrialMemberLimitBatch(ctx, members, 1)
}

// checkTrialMemberLimitBatch checks if adding `count` members would exceed the trial limit.
func (s *LocalService) checkTrialMemberLimitBatch(ctx context.Context, members []types.Member, count int) error {
	limit := s.d.Cfg.TrialMemberLimit
	if limit <= 0 {
		return nil // no limit configured
	}
	info, ok := ctxcompany.From(ctx)
	if !ok || info.Type != store.CompanyTypeTrial {
		return nil // not a trial company, no limit
	}
	if len(members)+count > limit {
		return domain.Forbidden(fmt.Sprintf("试用环境成员上限为 %d 人，升级后可扩容", limit))
	}
	return nil
}

// resolveOrCreateUser finds an existing user by phone or email, or creates a new one.
func (s *LocalService) resolveOrCreateUser(ctx context.Context, phone, email string) (uuid.UUID, error) {
	return core.ResolveOrCreateUser(ctx, s.d.Store, phone, email)
}
