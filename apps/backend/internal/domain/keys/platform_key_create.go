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
	platformKeys, err := s.store.Keys().PlatformKeys(ctx)
	if err != nil {
		return types.PlatformKey{}, err
	}
	groups, err := s.store.Budget().Groups(ctx)
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
		if input.Quota > budget.GetQuotaRemaining(members, platformKeys, *input.MemberID) {
			return types.PlatformKey{}, domain.Validation("额度不足，请先申请追加")
		}
	}

	var fullKeyPtr *string
	keyPrefix := "pending..."
	if s.relaySync == nil || !s.relaySync.Enabled() {
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
	if err := s.store.Keys().SetPlatformKeys(ctx, platformKeys); err != nil {
		return types.PlatformKey{}, err
	}

	if s.relaySync != nil && s.relaySync.Enabled() {
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
		if err := s.relaySync.SyncCreatePlatformKey(ctx, created, departmentID); err != nil {
			return types.PlatformKey{}, fmt.Errorf("relay sync enqueue: %w", err)
		}
		fullKey, err := s.relaySync.TrySyncCreate(ctx, created.ID)
		if err != nil {
			s.relaySync.RollbackFailedCreate(ctx, created.ID)
			return types.PlatformKey{}, domain.ServiceUnavailable("Relay 同步失败，请稍后重试")
		}
		_ = fullKey
		refreshed, err := s.store.Keys().PlatformKeys(ctx)
		if err != nil {
			return types.PlatformKey{}, err
		}
		for _, key := range refreshed {
			if key.ID == created.ID {
				return key, nil
			}
		}
	}
	return created, nil
}
