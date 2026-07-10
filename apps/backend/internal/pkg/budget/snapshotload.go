package budget

import (
	"context"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/pkg/common"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

func ResolveKeyPeriodKey(
	key types.PlatformKey,
	members []types.Member,
	groups []types.BudgetGroup,
	deptPeriod map[string]string,
	rootPeriodKey string,
	at time.Time,
) string {
	deptID := keyDepartmentID(key, members, groups)
	if deptID != "" {
		if orgPeriod, ok := deptPeriod[deptID]; ok {
			return SnapshotKey(orgPeriod, at)
		}
	}
	return rootPeriodKey
}

func ResolveGroupPeriodKeys(
	group types.BudgetGroup,
	deptPeriod map[string]string,
	rootPeriodKey string,
	at time.Time,
) []string {
	keys := make([]string, 0, len(group.DepartmentIDs))
	for _, deptID := range group.DepartmentIDs {
		if orgPeriod, ok := deptPeriod[deptID]; ok {
			keys = append(keys, SnapshotKey(orgPeriod, at))
		}
	}
	keys = uniqueStrings(keys)
	if len(keys) == 0 {
		return []string{rootPeriodKey}
	}
	return keys
}

func sumAxisConsumed(axisID string, periodKeys []string, byPeriod map[string]map[string]float64) float64 {
	var total float64
	for _, periodKey := range periodKeys {
		if consumedByAxis, ok := byPeriod[periodKey]; ok {
			total += consumedByAxis[axisID]
		}
	}
	return total
}

func PlatformKeyConsumed(
	ctx context.Context,
	snapshots store.BudgetSnapshotRepository,
	orgNodes store.OrgNodeRepository,
	key types.PlatformKey,
	members []types.Member,
	groups []types.BudgetGroup,
	clk clock.Clock,
) (float64, bool, error) {
	at := clock.NowUTC(clk)
	deptPeriod, rootPeriodKey, err := buildDeptPeriodMap(ctx, orgNodes, at)
	if err != nil {
		return 0, false, err
	}
	periodKey := ResolveKeyPeriodKey(key, members, groups, deptPeriod, rootPeriodKey, at)
	return snapshots.GetConsumed(ctx, store.SnapshotAxisPlatformKey, key.ID, periodKey)
}

func keyDepartmentID(key types.PlatformKey, members []types.Member, groups []types.BudgetGroup) string {
	if key.MemberID != nil {
		if member, ok := pkgorg.FindMemberByID(members, *key.MemberID); ok {
			return member.DepartmentID
		}
	}
	if key.BudgetGroupID != nil {
		for _, group := range groups {
			if group.ID == *key.BudgetGroupID && len(group.DepartmentIDs) > 0 {
				return group.DepartmentIDs[0]
			}
		}
	}
	return ""
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func mergeBudgetTreeConsumed(
	ctx context.Context,
	snapshots store.BudgetSnapshotRepository,
	tree []types.BudgetNode,
	clk clock.Clock,
) ([]types.BudgetNode, error) {
	at := clock.NowUTC(clk)
	var walk func(nodes []types.BudgetNode) error
	walk = func(nodes []types.BudgetNode) error {
		for i := range nodes {
			periodKey := SnapshotKey(nodes[i].Period, at)
			consumed, found, err := snapshots.GetConsumed(ctx, store.SnapshotAxisOrgNode, nodes[i].ID, periodKey)
			if err != nil {
				return err
			}
			if found {
				nodes[i].Consumed = consumed
			}
			if len(nodes[i].Children) > 0 {
				if err := walk(nodes[i].Children); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := walk(tree); err != nil {
		return nil, err
	}
	return tree, nil
}

func LoadPlatformKeysWithUsed(
	ctx context.Context,
	snapshots store.BudgetSnapshotRepository,
	org store.OrgRepository,
	budgetRepo store.BudgetRepository,
	keys store.KeysRepository,
	clk clock.Clock,
) ([]types.PlatformKey, error) {
	at := clock.NowUTC(clk)
	items, err := keys.PlatformKeys(ctx)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return items, nil
	}
	members, err := org.Members(ctx)
	if err != nil {
		return nil, err
	}
	groups, err := budgetRepo.Groups(ctx)
	if err != nil {
		return nil, err
	}
	deptPeriod, rootPeriodKey, err := buildDeptPeriodMap(ctx, org.Nodes(), at)
	if err != nil {
		return nil, err
	}
	keyPeriodKeys := make(map[string]string, len(items))
	periodKeys := make([]string, 0, len(items))
	for _, key := range items {
		periodKey := ResolveKeyPeriodKey(key, members, groups, deptPeriod, rootPeriodKey, at)
		keyPeriodKeys[key.ID] = periodKey
		periodKeys = append(periodKeys, periodKey)
	}
	byPeriod, err := snapshots.ListConsumedByPeriods(ctx, store.SnapshotAxisPlatformKey, uniqueStrings(periodKeys))
	if err != nil {
		return nil, err
	}
	usedByID := make(map[string]float64, len(items))
	for _, key := range items {
		periodKey := keyPeriodKeys[key.ID]
		if consumedByAxis, ok := byPeriod[periodKey]; ok {
			if used, ok := consumedByAxis[key.ID]; ok {
				usedByID[key.ID] = used
			}
		}
	}
	for i, key := range items {
		if used, ok := usedByID[key.ID]; ok {
			items[i].Used = used
		}
	}
	return items, nil
}

func LoadBudgetGroupsWithConsumed(
	ctx context.Context,
	snapshots store.BudgetSnapshotRepository,
	org store.OrgRepository,
	budgetRepo store.BudgetRepository,
	clk clock.Clock,
) ([]types.BudgetGroup, error) {
	at := clock.NowUTC(clk)
	groups, err := budgetRepo.Groups(ctx)
	if err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		return groups, nil
	}
	deptPeriod, rootPeriodKey, err := buildDeptPeriodMap(ctx, org.Nodes(), at)
	if err != nil {
		return nil, err
	}
	groupPeriodKeys := make(map[string][]string, len(groups))
	periodKeys := make([]string, 0, len(groups))
	for _, group := range groups {
		keys := ResolveGroupPeriodKeys(group, deptPeriod, rootPeriodKey, at)
		groupPeriodKeys[group.ID] = keys
		periodKeys = append(periodKeys, keys...)
	}
	byPeriod, err := snapshots.ListConsumedByPeriods(ctx, store.SnapshotAxisBudgetGroup, uniqueStrings(periodKeys))
	if err != nil {
		return nil, err
	}
	consumedByID := make(map[string]float64, len(groups))
	for _, group := range groups {
		consumedByID[group.ID] = sumAxisConsumed(group.ID, groupPeriodKeys[group.ID], byPeriod)
	}
	for i, group := range groups {
		if consumed, ok := consumedByID[group.ID]; ok {
			groups[i].Consumed = consumed
		}
	}
	return groups, nil
}

func LoadBudgetTreeWithConsumed(
	ctx context.Context,
	snapshots store.BudgetSnapshotRepository,
	orgNodes store.OrgNodeRepository,
	clk clock.Clock,
) ([]types.BudgetNode, error) {
	tree, err := common.LoadBudgetTree(ctx, orgNodes)
	if err != nil {
		return nil, err
	}
	return mergeBudgetTreeConsumed(ctx, snapshots, tree, clk)
}
