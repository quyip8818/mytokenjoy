package budget

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

func ResolveKeyPeriodKey(
	key types.PlatformKey,
	members []types.Member,
	projects []types.Project,
	deptPeriod map[uuid.UUID]string,
	rootPeriodKey string,
	at time.Time,
) string {
	deptID := keyDepartmentID(key, members, projects)
	if deptID != uuid.Nil {
		if orgPeriod, ok := deptPeriod[deptID]; ok {
			return SnapshotKey(orgPeriod, at)
		}
	}
	return rootPeriodKey
}

func sumAxisConsumed(axisID uuid.UUID, periodKeys []string, byPeriod map[string]map[uuid.UUID]float64) float64 {
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
	snapshots store.BudgetConsumedRepository,
	orgNodes store.OrgNodeRepository,
	key types.PlatformKey,
	members []types.Member,
	projects []types.Project,
	clk clock.Clock,
) (float64, bool, error) {
	at := clock.NowUTC(clk)
	deptPeriod, rootPeriodKey, err := buildDeptPeriodMap(ctx, orgNodes, at)
	if err != nil {
		return 0, false, err
	}
	periodKey := ResolveKeyPeriodKey(key, members, projects, deptPeriod, rootPeriodKey, at)
	return snapshots.GetConsumed(ctx, store.AxisKindPlatformKey, key.ID, periodKey)
}

func keyDepartmentID(key types.PlatformKey, members []types.Member, projects []types.Project) uuid.UUID {
	if key.MemberID != nil {
		if member, ok := pkgorg.FindMemberByID(members, *key.MemberID); ok {
			return member.DepartmentID
		}
	}
	if key.ProjectID != nil {
		for _, project := range projects {
			if project.ID == *key.ProjectID && project.OwnerDepartmentID != uuid.Nil {
				return project.OwnerDepartmentID
			}
		}
	}
	return uuid.Nil
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

func uniqueUUIDs(values []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(values))
	out := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		if value == uuid.Nil {
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

func LoadPlatformKeysWithUsed(
	ctx context.Context,
	snapshots store.BudgetConsumedRepository,
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
	projects, err := budgetRepo.Projects(ctx)
	if err != nil {
		return nil, err
	}
	deptPeriod, rootPeriodKey, err := buildDeptPeriodMap(ctx, org.Nodes(), at)
	if err != nil {
		return nil, err
	}
	keyPeriodKeys := make(map[uuid.UUID]string, len(items))
	periodKeys := make([]string, 0, len(items))
	for _, key := range items {
		periodKey := ResolveKeyPeriodKey(key, members, projects, deptPeriod, rootPeriodKey, at)
		keyPeriodKeys[key.ID] = periodKey
		periodKeys = append(periodKeys, periodKey)
	}
	byPeriod, err := snapshots.ListConsumedByPeriods(ctx, store.AxisKindPlatformKey, uniqueStrings(periodKeys))
	if err != nil {
		return nil, err
	}
	usedByID := make(map[uuid.UUID]float64, len(items))
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
			items[i].Consumed = used
		}
	}
	return items, nil
}

func LoadProjectsWithConsumed(
	ctx context.Context,
	snapshots store.BudgetConsumedRepository,
	org store.OrgRepository,
	budgetRepo store.BudgetRepository,
	clk clock.Clock,
) ([]types.Project, error) {
	at := clock.NowUTC(clk)
	projects, err := budgetRepo.Projects(ctx)
	if err != nil {
		return nil, err
	}
	if len(projects) == 0 {
		return projects, nil
	}
	members, err := org.Members(ctx)
	if err != nil {
		return nil, err
	}
	deptPeriod, rootPeriodKey, err := buildDeptPeriodMap(ctx, org.Nodes(), at)
	if err != nil {
		return nil, err
	}
	projectPeriodKeys := make(map[uuid.UUID][]string, len(projects))
	periodKeys := make([]string, 0, len(projects))
	for _, project := range projects {
		keys := ResolveProjectPeriodKeys(project, members, deptPeriod, rootPeriodKey, at)
		projectPeriodKeys[project.ID] = keys
		periodKeys = append(periodKeys, keys...)
	}
	byPeriod, err := snapshots.ListConsumedByPeriods(ctx, store.AxisKindProject, uniqueStrings(periodKeys))
	if err != nil {
		return nil, err
	}
	consumedByID := make(map[uuid.UUID]float64, len(projects))
	for _, project := range projects {
		consumedByID[project.ID] = sumAxisConsumed(project.ID, projectPeriodKeys[project.ID], byPeriod)
	}
	for i, project := range projects {
		if consumed, ok := consumedByID[project.ID]; ok {
			projects[i].Consumed = consumed
		}
	}
	return projects, nil
}

