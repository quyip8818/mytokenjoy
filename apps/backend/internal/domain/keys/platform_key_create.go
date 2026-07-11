package keys

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

func (s *service) CreatePlatformKey(ctx context.Context, input types.CreatePlatformKeyInput) (types.PlatformKey, error) {
	if err := s.delayer.Wait(ctx, 500*time.Millisecond); err != nil {
		return types.PlatformKey{}, err
	}
	members, err := s.store.Org().Members(ctx)
	if err != nil {
		return types.PlatformKey{}, err
	}
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
	platformKeys, err := budget.LoadPlatformKeysWithUsed(ctx, s.store.BudgetSnapshots(), s.store.Org(), s.store.Budget(), s.store.Keys(), s.cfg.Clock())
	if err != nil {
		return types.PlatformKey{}, err
	}
	groups, err := budget.LoadBudgetGroupsWithConsumed(ctx, s.store.BudgetSnapshots(), s.store.Org(), s.store.Budget(), s.cfg.Clock())
	if err != nil {
		return types.PlatformKey{}, err
	}

	if input.BudgetGroupID != nil {
		var group *types.BudgetGroup
		for i := range groups {
			if groups[i].ID == *input.BudgetGroupID {
				group = &groups[i]
				break
			}
		}
		if group == nil {
			return types.PlatformKey{}, domain.NotFound("Budget group not found")
		}
		if msg := budget.ValidateGroupKeyBudget(*group, platformKeys, input.Budget, ""); msg != nil {
			return types.PlatformKey{}, domain.Validation(*msg)
		}
		if input.MemberID != nil {
			if msg := common.ValidateModelIDsForMember(*input.MemberID, input.ModelWhitelist, members, departments, rules, models, common.ModelNotInDeptMessage); msg != nil {
				return types.PlatformKey{}, domain.Validation(*msg)
			}
		}
	} else {
		if input.MemberID == nil {
			return types.PlatformKey{}, domain.BadRequest("memberId required")
		}
		if msg := common.ValidateModelIDsForMember(*input.MemberID, input.ModelWhitelist, members, departments, rules, models, common.ModelNotInDeptMessage); msg != nil {
			return types.PlatformKey{}, domain.Validation(*msg)
		}
		if input.Budget > budget.GetBudgetRemaining(members, platformKeys, *input.MemberID) {
			return types.PlatformKey{}, domain.Validation("额度不足，请先申请追加")
		}
	}

	if err := s.requireNewAPI(); err != nil {
		return types.PlatformKey{}, err
	}

	created := types.PlatformKey{
		ID:   fmt.Sprintf("plk-%d", time.Now().UnixMilli()),
		Name: input.Name, KeyPrefix: "pending...", MemberID: input.MemberID,
		BudgetGroupID: input.BudgetGroupID,
		Status:        "active", Budget: input.Budget, Used: 0,
		ModelWhitelist: append([]int64{}, input.ModelWhitelist...),
		CreatedAt:      time.Now().Format("2006-01-02"),
	}
	platformKeys = append(platformKeys, created)
	if err := s.store.Keys().SetPlatformKeys(ctx, platformKeys); err != nil {
		return types.PlatformKey{}, err
	}

	departmentID, err := s.resolvePlatformKeyDepartmentID(input, members, groups)
	if err != nil {
		return types.PlatformKey{}, err
	}
	return s.syncPlatformKeyCreate(ctx, created, departmentID)
}
