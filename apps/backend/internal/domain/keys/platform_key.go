package keys

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/org"
)

func (s *service) CreatePlatformKey(ctx context.Context, input types.CreatePlatformKeyInput) (types.PlatformKey, error) {
	if err := s.delayer.Wait(ctx, 500*time.Millisecond); err != nil {
		return types.PlatformKey{}, err
	}
	members := s.store.Org().Members()
	departments := s.store.Org().Departments()
	rules := s.store.Models().RoutingRules()
	models := s.store.Models().Models()
	pools := s.store.Budget().MemberQuotaPools()
	platformKeys := s.store.Keys().PlatformKeys()
	groups := s.store.Budget().Groups()

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
		if msg := budget.ValidateGroupKeyQuota(*group, platformKeys, input.Quota, ""); msg != nil {
			return types.PlatformKey{}, domain.Validation(*msg)
		}
		if input.MemberID != nil {
			if msg := common.ValidateModelsForMember(*input.MemberID, input.ModelWhitelist, members, departments, rules, models, common.ModelNotInDeptMessage); msg != nil {
				return types.PlatformKey{}, domain.Validation(*msg)
			}
		}
	} else {
		if input.MemberID == nil {
			return types.PlatformKey{}, domain.BadRequest("memberId required")
		}
		if msg := common.ValidateModelsForMember(*input.MemberID, input.ModelWhitelist, members, departments, rules, models, common.ModelNotInDeptMessage); msg != nil {
			return types.PlatformKey{}, domain.Validation(*msg)
		}
		if input.Quota > budget.GetQuotaRemaining(pools, platformKeys, *input.MemberID) {
			return types.PlatformKey{}, domain.Validation("额度不足，请先申请追加")
		}
	}

	var fullKeyPtr *string
	keyPrefix := "pending..."
	if s.lifecycle == nil || !s.lifecycle.Enabled() {
		fullKey := fmt.Sprintf("%s%d-demo-secret-key", common.DemoKeyPrefix, time.Now().UnixMilli())
		fullKeyPtr = &fullKey
		keyPrefix = fullKey
		if len(keyPrefix) > 12 {
			keyPrefix = keyPrefix[:12] + "..."
		}
	}
	memberName := (*string)(nil)
	if input.MemberID != nil {
		if member, ok := org.FindMemberByID(members, *input.MemberID); ok {
			memberName = &member.Name
		}
	}
	var groupName *string
	if input.BudgetGroupID != nil {
		for _, group := range groups {
			if group.ID == *input.BudgetGroupID {
				groupName = &group.Name
				break
			}
		}
	}
	created := types.PlatformKey{
		ID:   fmt.Sprintf("plk-%d", time.Now().UnixMilli()),
		Name: input.Name, KeyPrefix: keyPrefix, FullKey: fullKeyPtr,
		MemberID: input.MemberID, MemberName: memberName, AppName: input.AppName,
		BudgetGroupID: input.BudgetGroupID, BudgetGroupName: groupName,
		Status: "active", Quota: input.Quota, Used: 0,
		ModelWhitelist: append([]string{}, input.ModelWhitelist...),
		CreatedAt:      time.Now().Format("2006-01-02"),
	}
	platformKeys = append(platformKeys, created)
	if err := s.store.Keys().SetPlatformKeys(platformKeys); err != nil {
		return types.PlatformKey{}, err
	}

	if s.lifecycle != nil && s.lifecycle.Enabled() {
		departmentID := ""
		if input.MemberID != nil {
			if member, ok := org.FindMemberByID(members, *input.MemberID); ok {
				departmentID = member.DepartmentID
			}
		}
		if departmentID == "" && input.BudgetGroupID != nil {
			for _, group := range groups {
				if group.ID == *input.BudgetGroupID && len(group.DepartmentIDs) > 0 {
					departmentID = group.DepartmentIDs[0]
					break
				}
			}
		}
		if departmentID == "" {
			return types.PlatformKey{}, domain.Validation("无法解析部门用于 Relay 同步")
		}
		if err := s.lifecycle.SyncCreatePlatformKey(ctx, created, departmentID); err != nil {
			return types.PlatformKey{}, fmt.Errorf("relay sync enqueue: %w", err)
		}
		fullKey, err := s.lifecycle.TrySyncCreate(ctx, created.ID)
		if err != nil {
			s.lifecycle.RollbackFailedCreate(created.ID)
			return types.PlatformKey{}, domain.ServiceUnavailable("Relay 同步失败，请稍后重试")
		}
		_ = fullKey
		for _, key := range s.store.Keys().PlatformKeys() {
			if key.ID == created.ID {
				return key, nil
			}
		}
	}
	return created, nil
}

func (s *service) UpdatePlatformKey(ctx context.Context, id string, input types.UpdatePlatformKeyInput) (types.PlatformKey, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.PlatformKey{}, err
	}
	platformKeys := s.store.Keys().PlatformKeys()
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
	members := s.store.Org().Members()
	departments := s.store.Org().Departments()
	rules := s.store.Models().RoutingRules()
	models := s.store.Models().Models()
	pools := s.store.Budget().MemberQuotaPools()
	groups := s.store.Budget().Groups()

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
			if otherAllocated+*input.Quota > budget.GetPersonalQuota(pools, *existing.MemberID) {
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
	if err := s.store.Keys().SetPlatformKeys(platformKeys); err != nil {
		return types.PlatformKey{}, err
	}
	if s.lifecycle != nil && s.lifecycle.Enabled() {
		if err := s.lifecycle.EnqueueUpdatePlatformKey(id); err != nil {
			return types.PlatformKey{}, err
		}
	}
	return existing, nil
}

func (s *service) TogglePlatformKey(ctx context.Context, id string, enabled bool) (types.PlatformKey, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.PlatformKey{}, err
	}
	platformKeys := s.store.Keys().PlatformKeys()
	for i := range platformKeys {
		if platformKeys[i].ID == id {
			if enabled {
				platformKeys[i].Status = "active"
			} else {
				platformKeys[i].Status = "disabled"
			}
			if err := s.store.Keys().SetPlatformKeys(platformKeys); err != nil {
				return types.PlatformKey{}, err
			}
			if s.lifecycle != nil && s.lifecycle.Enabled() {
				if err := s.lifecycle.EnqueueUpdatePlatformKey(id); err != nil {
					return types.PlatformKey{}, err
				}
			}
			return platformKeys[i], nil
		}
	}
	return types.PlatformKey{}, domain.NotFound("Not found")
}

func (s *service) RotatePlatformKey(ctx context.Context, id string) (types.PlatformKey, error) {
	if err := s.delayer.Wait(ctx, 500*time.Millisecond); err != nil {
		return types.PlatformKey{}, err
	}
	platformKeys := s.store.Keys().PlatformKeys()
	for i := range platformKeys {
		if platformKeys[i].ID == id {
			fullKey := fmt.Sprintf("tj-rot-%d-demo-secret", time.Now().UnixMilli())
			platformKeys[i].FullKey = &fullKey
			prefix := fullKey
			if len(prefix) > 12 {
				prefix = prefix[:12] + "..."
			}
			platformKeys[i].KeyPrefix = prefix
			if err := s.store.Keys().SetPlatformKeys(platformKeys); err != nil {
				return types.PlatformKey{}, err
			}
			return platformKeys[i], nil
		}
	}
	return types.PlatformKey{}, domain.NotFound("Not found")
}

func (s *service) RevokePlatformKey(ctx context.Context, id string) error {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return err
	}
	platformKeys := s.store.Keys().PlatformKeys()
	for i := range platformKeys {
		if platformKeys[i].ID == id {
			platformKeys[i].Status = "revoked"
			if err := s.store.Keys().SetPlatformKeys(platformKeys); err != nil {
				return err
			}
			if s.lifecycle != nil && s.lifecycle.Enabled() {
				return s.lifecycle.SyncRevokePlatformKey(ctx, id)
			}
			return nil
		}
	}
	return domain.NotFound("Not found")
}

func (s *service) DeletePlatformKey(id string) error {
	platformKeys := s.store.Keys().PlatformKeys()
	for i := range platformKeys {
		if platformKeys[i].ID == id {
			platformKeys = append(platformKeys[:i], platformKeys[i+1:]...)
			if err := s.store.Keys().SetPlatformKeys(platformKeys); err != nil {
				return err
			}
			return nil
		}
	}
	return domain.NotFound("Not found")
}
