package budget

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/store"
)

// MappingStores groups repositories needed to compute remain budget for a platform key mapping.
type MappingStores struct {
	Consumed store.BudgetConsumedRepository
	OrgNodes store.OrgNodeRepository
	Org      store.OrgRepository
	Budget   store.BudgetRepository
	Keys     store.KeysRepository
	Clock    clock.Clock
}

// RemainForMapping returns effective remaining budget for ingest cap checks.
func RemainForMapping(
	ctx context.Context,
	stores MappingStores,
	mapping *store.PlatformKeyMapping,
	periodKey string,
) (float64, error) {
	if mapping.DepartmentID == "" {
		return 0, fmt.Errorf("department not found")
	}
	budgetCtx, err := LoadBudgetContext(ctx, stores.Consumed, stores.Org, stores.Budget, stores.Keys, stores.Clock)
	if err != nil {
		return 0, err
	}
	return ComputeRemainForMapping(ctx, budgetCtx, stores.Consumed, stores.Org, *mapping, periodKey)
}

// ComputeRemainForMapping uses a preloaded budget context for batch gateway summary writes.
func ComputeRemainForMapping(
	ctx context.Context,
	budgetCtx BudgetContext,
	consumed store.BudgetConsumedRepository,
	org store.OrgRepository,
	mapping store.PlatformKeyMapping,
	periodKey string,
) (float64, error) {
	if mapping.DepartmentID == "" {
		return 0, fmt.Errorf("department not found")
	}
	limit, found, err := org.Nodes().GetNodeBudget(ctx, mapping.DepartmentID)
	if err != nil {
		return 0, err
	}
	if !found || limit <= 0 {
		return 0, fmt.Errorf("budget exceeded")
	}

	key, ok := budgetCtx.FindPlatformKey(mapping.PlatformKeyID)
	if !ok {
		return 0, fmt.Errorf("platform key not found")
	}
	if key.Budget > 0 {
		keyUsed, found, err := consumed.GetConsumed(ctx, store.AxisKindPlatformKey, key.ID, periodKey)
		if err != nil {
			return 0, err
		}
		if found {
			key.Used = keyUsed
		} else {
			key.Used = 0
		}
	}

	deptConsumed, _, err := consumed.GetConsumed(ctx, store.AxisKindOrgNode, mapping.DepartmentID, periodKey)
	if err != nil {
		return 0, err
	}
	deptAxis := &DeptAxisInput{Budget: limit, Consumed: deptConsumed}

	var memberAxis *MemberAxisInput
	if mapping.MemberID != nil && key.BudgetGroupID == nil {
		quota, memberFound, err := org.MemberPersonalBudget(ctx, *mapping.MemberID)
		if err != nil {
			return 0, err
		}
		if !memberFound {
			memberAxis = &MemberAxisInput{Skip: true}
		} else {
			memberConsumed, _, err := consumed.GetConsumed(ctx, store.AxisKindMember, *mapping.MemberID, periodKey)
			if err != nil {
				return 0, err
			}
			memberAxis = &MemberAxisInput{Cap: quota, Consumed: memberConsumed}
		}
	}

	return budgetCtx.ComputeRemain(key, mapping.DepartmentID, memberAxis, deptAxis), nil
}
