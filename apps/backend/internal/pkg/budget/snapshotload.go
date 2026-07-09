package budget

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

func BuildDeptPeriodMap(ctx context.Context, nodes store.OrgNodeRepository, at time.Time) (map[string]string, string, error) {
	orgNodes, err := nodes.Tree(ctx)
	if err != nil {
		return nil, "", err
	}
	flat := pkgorg.FlattenOrgNodeTree(orgNodes)
	deptPeriod := make(map[string]string, len(flat))
	var rootPeriodKey string
	for _, node := range flat {
		deptPeriod[node.ID] = node.Period
		if node.ParentID == nil || *node.ParentID == "" {
			rootPeriodKey = SnapshotKey(node.Period, at)
		}
	}
	if rootPeriodKey == "" {
		return nil, "", fmt.Errorf("org tree has no root node")
	}
	return deptPeriod, rootPeriodKey, nil
}

func DepartmentPeriodKey(ctx context.Context, nodes store.OrgNodeRepository, departmentID string, at time.Time) (string, error) {
	if departmentID == "" {
		return SnapshotKey(PeriodMonthly, at), nil
	}
	orgPeriod, found, err := nodes.GetNodePeriod(ctx, departmentID)
	if err != nil {
		return "", err
	}
	if !found {
		return SnapshotKey(PeriodMonthly, at), nil
	}
	return SnapshotKey(orgPeriod, at), nil
}

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
	at time.Time,
) (float64, bool, error) {
	deptPeriod, rootPeriodKey, err := BuildDeptPeriodMap(ctx, orgNodes, at)
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

func MergeBudgetTreeConsumed(
	ctx context.Context,
	snapshots store.BudgetSnapshotRepository,
	tree []types.BudgetNode,
	at time.Time,
) ([]types.BudgetNode, error) {
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
	at time.Time,
) ([]types.PlatformKey, error) {
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
	deptPeriod, rootPeriodKey, err := BuildDeptPeriodMap(ctx, org.Nodes(), at)
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
	at time.Time,
) ([]types.BudgetGroup, error) {
	groups, err := budgetRepo.Groups(ctx)
	if err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		return groups, nil
	}
	deptPeriod, rootPeriodKey, err := BuildDeptPeriodMap(ctx, org.Nodes(), at)
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
	at time.Time,
) ([]types.BudgetNode, error) {
	tree, err := common.LoadBudgetTree(ctx, orgNodes)
	if err != nil {
		return nil, err
	}
	return MergeBudgetTreeConsumed(ctx, snapshots, tree, at)
}
