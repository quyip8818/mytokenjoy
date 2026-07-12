package keys

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

func (s *service) CreatePlatformKey(ctx context.Context, input types.CreatePlatformKeyInput) (types.PlatformKey, error) {
	if err := s.delayer.Wait(ctx, 500*time.Millisecond); err != nil {
		return types.PlatformKey{}, err
	}
	if input.Scope == "" {
		return types.PlatformKey{}, domain.Validation("scope is required")
	}
	if err := budget.ValidatePlatformKeyScope(input.Scope, input.MemberID, input.ProjectID); err != nil {
		return types.PlatformKey{}, domain.Validation(err.Error())
	}

	budgetCtx, err := budget.LoadBudgetContext(ctx, s.store.BudgetConsumed(), s.store.Org(), s.store.Budget(), s.store.Keys(), s.cfg.Clock())
	if err != nil {
		return types.PlatformKey{}, err
	}
	platformKeys := budgetCtx.PlatformKeys
	projects := budgetCtx.Projects
	members := budgetCtx.Members
	departments, err := common.LoadDepartments(ctx, s.store.Org().Nodes())
	if err != nil {
		return types.PlatformKey{}, err
	}
	rules, err := common.LoadRoutingRules(ctx, s.store.Org().Nodes(), s.store.Models().Allowlist())
	if err != nil {
		return types.PlatformKey{}, err
	}
	models, err := s.store.Models().Models(ctx)
	if err != nil {
		return types.PlatformKey{}, err
	}

	switch input.Scope {
	case types.PlatformKeyScopeMember:
		if msg := common.ValidateModelIDsForMember(*input.MemberID, input.ModelWhitelist, members, departments, rules, models, common.ModelNotInDeptMessage); msg != nil {
			return types.PlatformKey{}, domain.Validation(*msg)
		}
		if msg := budget.ValidateMemberScopeKeyBudget(members, platformKeys, *input.MemberID, input.Budget, ""); msg != nil {
			return types.PlatformKey{}, domain.Validation(*msg)
		}
	case types.PlatformKeyScopeProject:
		project, ok := budget.FindProject(projects, *input.ProjectID)
		if !ok {
			return types.PlatformKey{}, domain.NotFound("Project not found")
		}
		if msg := budget.ValidateProjectScopeKeyBudget(input.Scope, project, platformKeys, input.MemberID, input.Budget, ""); msg != nil {
			return types.PlatformKey{}, domain.Validation(*msg)
		}
		if input.MemberID != nil {
			if msg := common.ValidateModelIDsForMember(*input.MemberID, input.ModelWhitelist, members, departments, rules, models, common.ModelNotInDeptMessage); msg != nil {
				return types.PlatformKey{}, domain.Validation(*msg)
			}
		}
	case types.PlatformKeyScopeProjectMember:
		project, ok := budget.FindProject(projects, *input.ProjectID)
		if !ok {
			return types.PlatformKey{}, domain.NotFound("Project not found")
		}
		if err := budget.ValidateProjectMemberRoster(project, *input.MemberID); err != nil {
			return types.PlatformKey{}, domain.Validation(err.Error())
		}
		if msg := common.ValidateModelIDsForMember(*input.MemberID, input.ModelWhitelist, members, departments, rules, models, common.ModelNotInDeptMessage); msg != nil {
			return types.PlatformKey{}, domain.Validation(*msg)
		}
		if msg := budget.ValidateProjectScopeKeyBudget(input.Scope, project, platformKeys, input.MemberID, input.Budget, ""); msg != nil {
			return types.PlatformKey{}, domain.Validation(*msg)
		}
	}

	if err := s.requireNewAPI(); err != nil {
		return types.PlatformKey{}, err
	}

	created := types.PlatformKey{
		ID:   fmt.Sprintf("plk-%d", time.Now().UnixMilli()),
		Name: input.Name, KeyPrefix: "pending...", Scope: input.Scope,
		MemberID: input.MemberID, ProjectID: input.ProjectID,
		Status: "active", Budget: input.Budget, Consumed: 0,
		ModelWhitelist: append([]int64{}, input.ModelWhitelist...),
		CreatedAt:      time.Now().Format("2006-01-02"),
	}
	platformKeys = append(platformKeys, created)
	if err := s.store.Keys().SetPlatformKeys(ctx, platformKeys); err != nil {
		return types.PlatformKey{}, err
	}

	departmentID, err := s.resolvePlatformKeyDepartmentID(input, members, projects)
	if err != nil {
		return types.PlatformKey{}, err
	}
	result, err := s.syncPlatformKeyCreate(ctx, created, departmentID)
	if err != nil {
		return types.PlatformKey{}, err
	}
	if err := domainbudget.RefreshPlatformKeySoft(ctx, s.store, created.ID, s.cfg.Clock(), nil); err != nil {
		return types.PlatformKey{}, err
	}
	return result, nil
}
