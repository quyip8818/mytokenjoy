package keys

import (
	"context"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

func (s *service) UpdatePlatformKey(ctx context.Context, id string, input types.UpdatePlatformKeyInput) (types.PlatformKey, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.PlatformKey{}, err
	}
	platformKeys, err := s.store.Keys().PlatformKeys(ctx)
	if err != nil {
		return types.PlatformKey{}, err
	}
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
	groups, err := s.store.Budget().Groups(ctx)
	if err != nil {
		return types.PlatformKey{}, err
	}

	if len(input.ModelWhitelist) > 0 && existing.MemberID != nil {
		if msg := common.ValidateModelsForMember(*existing.MemberID, input.ModelWhitelist, members, departments, rules, models, common.ModelNotInDeptMessage); msg != nil {
			return types.PlatformKey{}, domain.Validation(*msg)
		}
	}
	if input.Quota != nil {
		if existing.BudgetGroupID != nil {
			var group *types.BudgetGroup
			for i := range groups {
				if groups[i].ID == *existing.BudgetGroupID {
					group = &groups[i]
					break
				}
			}
			if group == nil {
				return types.PlatformKey{}, domain.NotFound("Budget group not found")
			}
			if msg := budget.ValidateGroupKeyQuota(*group, platformKeys, *input.Quota, existing.ID); msg != nil {
				return types.PlatformKey{}, domain.Validation(*msg)
			}
		} else if existing.MemberID != nil {
			otherAllocated := 0.0
			for _, key := range platformKeys {
				if key.MemberID != nil && *key.MemberID == *existing.MemberID && key.BudgetGroupID == nil && key.Status == "active" && key.ID != existing.ID {
					otherAllocated += key.Quota
				}
			}
			if otherAllocated+*input.Quota > budget.GetPersonalQuota(members, *existing.MemberID) {
				return types.PlatformKey{}, domain.Validation("额度不足，请先申请追加")
			}
		}
	}
	if input.Name != nil {
		existing.Name = *input.Name
	}
	if input.Quota != nil {
		existing.Quota = *input.Quota
	}
	if input.ModelWhitelist != nil {
		existing.ModelWhitelist = append([]string{}, input.ModelWhitelist...)
	}
	platformKeys[idx] = existing
	if err := s.store.Keys().SetPlatformKeys(ctx, platformKeys); err != nil {
		return types.PlatformKey{}, err
	}
	if s.relaySync != nil && s.relaySync.Enabled() {
		if err := s.relaySync.EnqueueUpdatePlatformKey(ctx, id); err != nil {
			return types.PlatformKey{}, err
		}
	}
	return existing, nil
}
