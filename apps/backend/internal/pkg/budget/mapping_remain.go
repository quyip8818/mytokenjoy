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
// periodKey is the open budget period for the mapping's department.
func RemainForMapping(
	ctx context.Context,
	stores MappingStores,
	mapping *store.PlatformKeyMapping,
	periodKey string,
) (float64, error) {
	if mapping.DepartmentID == "" {
		return 0, fmt.Errorf("department not found")
	}
	limit, found, err := stores.OrgNodes.GetNodeBudget(ctx, mapping.DepartmentID)
	if err != nil {
		return 0, err
	}
	if !found || limit <= 0 {
		return 0, fmt.Errorf("budget exceeded")
	}

	budgetCtx, err := LoadBudgetContext(ctx, stores.Consumed, stores.Org, stores.Budget, stores.Keys, stores.Clock)
	if err != nil {
		return 0, err
	}
	key, ok := budgetCtx.FindPlatformKey(mapping.PlatformKeyID)
	if !ok {
		return 0, fmt.Errorf("platform key not found")
	}
	if key.Budget > 0 {
		consumed, found, err := stores.Consumed.GetConsumed(ctx, store.AxisKindPlatformKey, key.ID, periodKey)
		if err != nil {
			return 0, err
		}
		if found {
			key.Used = consumed
		} else {
			key.Used = 0
		}
	}

	deptConsumed, _, err := stores.Consumed.GetConsumed(ctx, store.AxisKindOrgNode, mapping.DepartmentID, periodKey)
	if err != nil {
		return 0, err
	}
	deptAxis := &DeptAxisInput{Budget: limit, Consumed: deptConsumed}

	var memberAxis *MemberAxisInput
	if mapping.MemberID != nil && key.BudgetGroupID == nil {
		quota, memberFound, err := stores.Org.MemberPersonalBudget(ctx, *mapping.MemberID)
		if err != nil {
			return 0, err
		}
		if !memberFound {
			memberAxis = &MemberAxisInput{Skip: true}
		} else {
			memberConsumed, _, err := stores.Consumed.GetConsumed(ctx, store.AxisKindMember, *mapping.MemberID, periodKey)
			if err != nil {
				return 0, err
			}
			memberAxis = &MemberAxisInput{Cap: quota, Consumed: memberConsumed}
		}
	}

	return budgetCtx.ComputeRemain(key, mapping.DepartmentID, memberAxis, deptAxis), nil
}
