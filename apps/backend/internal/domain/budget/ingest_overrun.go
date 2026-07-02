package budget

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *IngestService) evaluateOverrun(
	ctx context.Context,
	st store.Store,
	tree []types.BudgetNode,
	members []types.Member,
	memberID *string,
	mapping *store.RelayMapping,
) error {
	if s.lifecycle == nil || !s.lifecycle.Enabled() {
		return nil
	}
	platformKeys, err := st.Keys().PlatformKeys(ctx)
	if err != nil {
		return err
	}

	if memberID != nil && mapping.BudgetGroupID == nil {
		used := pkgbudget.GetUsedKeyQuota(platformKeys, *memberID)
		capacity := pkgbudget.GetPersonalQuota(members, *memberID)
		if used >= capacity {
			s.notifyOverrun(ctx, types.NotificationEventOverrunBlocked, *memberID, map[string]any{
				"scope": "member", "memberId": *memberID, "used": used, "capacity": capacity,
			})
			return s.disableMemberKeys(ctx, platformKeys, *memberID)
		}
	}

	if node := pkgbudget.FindBudgetNode(tree, mapping.DepartmentID); node != nil {
		if node.Consumed >= node.Budget {
			s.notifyOverrun(ctx, types.NotificationEventOverrunBlocked, mapping.DepartmentID, map[string]any{
				"scope": "department", "departmentId": mapping.DepartmentID,
				"consumed": node.Consumed, "budget": node.Budget,
			})
			return s.disableDepartmentKeys(ctx, mapping.DepartmentID)
		}
	}

	if mapping.BudgetGroupID != nil {
		groups, err := st.Budget().Groups(ctx)
		if err != nil {
			return err
		}
		for _, group := range groups {
			if group.ID == *mapping.BudgetGroupID && group.Consumed >= group.Budget {
				s.notifyOverrun(ctx, types.NotificationEventOverrunBlocked, group.ID, map[string]any{
					"scope": "budgetGroup", "budgetGroupId": group.ID,
					"consumed": group.Consumed, "budget": group.Budget,
				})
				return s.disableBudgetGroupKeys(ctx, *mapping.BudgetGroupID)
			}
		}
	}
	return nil
}

func (s *IngestService) notifyOverrun(ctx context.Context, eventType, recipient string, payload map[string]any) {
	if s.notifier == nil {
		return
	}
	_ = s.notifier.Send(ctx, types.Notification{EventType: eventType, Recipient: recipient, Payload: payload})
}

func (s *IngestService) disableMemberKeys(ctx context.Context, platformKeys []types.PlatformKey, memberID string) error {
	for _, key := range platformKeys {
		if key.MemberID != nil && *key.MemberID == memberID && key.BudgetGroupID == nil && key.Status == "active" {
			if err := s.lifecycle.DisablePlatformKey(ctx, key.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *IngestService) disableDepartmentKeys(ctx context.Context, departmentID string) error {
	mappings, err := s.store.Relay().ListMappingsByDepartmentID(ctx, departmentID)
	if err != nil {
		return err
	}
	for _, mapping := range mappings {
		if err := s.lifecycle.DisablePlatformKey(ctx, mapping.PlatformKeyID); err != nil {
			return err
		}
	}
	return nil
}

func (s *IngestService) disableBudgetGroupKeys(ctx context.Context, budgetGroupID string) error {
	mappings, err := s.store.Relay().ListMappingsByBudgetGroupID(ctx, budgetGroupID)
	if err != nil {
		return err
	}
	for _, mapping := range mappings {
		if err := s.lifecycle.DisablePlatformKey(ctx, mapping.PlatformKeyID); err != nil {
			return err
		}
	}
	return nil
}
