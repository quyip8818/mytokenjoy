package keys

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

func (s *service) UpdatePlatformKey(ctx context.Context, id uuid.UUID, input types.UpdatePlatformKeyInput) (types.PlatformKey, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.PlatformKey{}, err
	}
	budgetCtx, err := budget.LoadBudgetContext(ctx, s.store.BudgetConsumed(), s.store.Org(), s.store.Budget(), s.store.Keys(), s.cfg.Clock())
	if err != nil {
		return types.PlatformKey{}, err
	}
	platformKeys := budgetCtx.PlatformKeys
	projects := budgetCtx.Projects
	idx := -1
	for i := range platformKeys {
		if platformKeys[i].ID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		return types.PlatformKey{}, domain.NotFound("Not found")
	}
	existing := platformKeys[idx]
	previous := existing
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

	if len(input.ModelWhitelist) > 0 && existing.MemberID != nil {
		if msg := common.ValidateModelIDsForMember(*existing.MemberID, input.ModelWhitelist, members, departments, rules, models, common.ModelNotInDeptMessage); msg != nil {
			return types.PlatformKey{}, domain.Validation(*msg)
		}
	}
	if input.Budget != nil {
		switch existing.Scope {
		case types.PlatformKeyScopeMember:
			if existing.MemberID == nil {
				return types.PlatformKey{}, domain.Validation("memberId required for member scope")
			}
			if msg := budget.ValidateMemberScopeKeyBudget(members, platformKeys, *existing.MemberID, int64(*input.Budget), existing.ID); msg != nil {
				return types.PlatformKey{}, domain.Validation(*msg)
			}
		case types.PlatformKeyScopeProject, types.PlatformKeyScopeProjectMember:
			if existing.ProjectID == nil {
				return types.PlatformKey{}, domain.Validation("projectId required")
			}
			project, ok := budget.FindProject(projects, *existing.ProjectID)
			if !ok {
				return types.PlatformKey{}, domain.NotFound("Project not found")
			}
			if msg := budget.ValidateProjectScopeKeyBudget(existing.Scope, project, platformKeys, existing.MemberID, int64(*input.Budget), existing.ID); msg != nil {
				return types.PlatformKey{}, domain.Validation(*msg)
			}
		default:
			return types.PlatformKey{}, domain.Validation("invalid platform key scope")
		}
	}
	if input.Name != nil {
		existing.Name = *input.Name
	}
	if input.Budget != nil {
		existing.Budget = int64(*input.Budget)
	}
	if input.ModelWhitelist != nil {
		existing.ModelWhitelist = append([]uuid.UUID{}, input.ModelWhitelist...)
	}
	if err := s.requireNewAPI(); err != nil {
		return types.PlatformKey{}, err
	}
	return s.persistPlatformKeyWithNewAPISync(ctx, platformKeys, idx, existing, previous, id)
}
