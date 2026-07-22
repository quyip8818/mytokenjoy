package budget

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

func (s *service) GetTree(ctx context.Context) ([]types.BudgetNode, error) {
	return common.LoadBudgetTree(ctx, s.store.Org().Nodes())
}

func (s *service) MemberSummary(ctx context.Context, memberID uuid.UUID) (types.MemberBudgetSummary, error) {
	if memberID == uuid.Nil {
		return types.MemberBudgetSummary{}, domain.BadRequest("memberId is required")
	}
	budgetCtx, err := pkgbudget.LoadBudgetContext(ctx, s.store.BudgetConsumed(), s.store.Org(), s.store.Budget(), s.store.Keys(), s.cfg.Clock())
	if err != nil {
		return types.MemberBudgetSummary{}, err
	}
	reservedPool := pkgbudget.GetReservedPoolForMember(budgetCtx.Tree, budgetCtx.Members, memberID)
	return pkgbudget.BuildBudgetSummary(budgetCtx.Members, budgetCtx.PlatformKeys, memberID, reservedPool), nil
}

func (s *service) ListMemberBudgets(ctx context.Context, deptID uuid.UUID) ([]types.MemberBudget, error) {
	budgetCtx, err := pkgbudget.LoadBudgetContext(ctx, s.store.BudgetConsumed(), s.store.Org(), s.store.Budget(), s.store.Keys(), s.cfg.Clock())
	if err != nil {
		return nil, err
	}
	if pkgbudget.FindBudgetNode(budgetCtx.Tree, deptID) == nil {
		return nil, domain.NotFound("Department not found")
	}
	memberBudgets := make([]types.MemberBudget, 0)
	for _, member := range budgetCtx.Members {
		if member.DepartmentID == deptID {
			memberBudgets = append(memberBudgets, pkgbudget.BuildMemberBudget(member, budgetCtx.PlatformKeys))
		}
	}
	return memberBudgets, nil
}

func (s *service) GetProjectMemberConsumed(ctx context.Context, projectID uuid.UUID) (map[uuid.UUID]int64, error) {
	projects, err := s.store.Budget().Projects(ctx)
	if err != nil {
		return nil, err
	}
	var target *types.Project
	for i := range projects {
		if projects[i].ID == projectID {
			target = &projects[i]
			break
		}
	}
	if target == nil {
		return nil, domain.NotFound("Project not found")
	}
	if len(target.MemberIDs) == 0 {
		return make(map[uuid.UUID]int64), nil
	}

	deptID := target.OwnerDepartmentID
	open, err := pkgbudget.OpenDepartmentPeriod(ctx, s.store.Org().Nodes(), deptID, s.cfg.Clock())
	if err != nil {
		return nil, err
	}
	periodKey := open.String()

	keys, err := s.store.Keys().PlatformKeys(ctx)
	if err != nil {
		return nil, err
	}
	consumedRepo := s.store.BudgetConsumed()

	result := make(map[uuid.UUID]int64, len(target.MemberIDs))
	for _, memberID := range target.MemberIDs {
		sum, err := pkgbudget.SumProjectMemberKeyConsumedFromRepo(
			ctx, consumedRepo, keys, projectID, memberID, periodKey,
		)
		if err != nil {
			return nil, err
		}
		result[memberID] = sum
	}
	return result, nil
}
