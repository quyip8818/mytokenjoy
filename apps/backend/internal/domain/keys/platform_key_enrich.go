package keys

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type platformKeyLookups struct {
	memberByID map[string]types.Member
	groupByID  map[string]types.BudgetGroup
}

func (s *service) loadPlatformKeyLookups(ctx context.Context) (platformKeyLookups, error) {
	members, err := s.store.Org().Members(ctx)
	if err != nil {
		return platformKeyLookups{}, err
	}
	memberByID := make(map[string]types.Member, len(members))
	for _, member := range members {
		memberByID[member.ID] = member
	}
	groups, err := s.store.Budget().Groups(ctx)
	if err != nil {
		return platformKeyLookups{}, err
	}
	groupByID := make(map[string]types.BudgetGroup, len(groups))
	for _, group := range groups {
		groupByID[group.ID] = group
	}
	return platformKeyLookups{memberByID: memberByID, groupByID: groupByID}, nil
}

func enrichPlatformKey(key types.PlatformKey, lookups platformKeyLookups) types.PlatformKey {
	enriched := key
	enriched.MemberName = nil
	enriched.BudgetGroupName = nil
	enriched.ProjectName = nil

	if key.BudgetGroupID != nil {
		enriched.Type = "project"
		if group, ok := lookups.groupByID[*key.BudgetGroupID]; ok {
			name := group.Name
			enriched.BudgetGroupName = &name
		}
		switch {
		case enriched.BudgetGroupName != nil && *enriched.BudgetGroupName != "":
			name := *enriched.BudgetGroupName
			enriched.ProjectName = &name
		case key.AppName != nil && *key.AppName != "":
			name := *key.AppName
			enriched.ProjectName = &name
		}
	} else {
		enriched.Type = "member"
	}

	if key.MemberID != nil {
		if member, ok := lookups.memberByID[*key.MemberID]; ok {
			name := member.Name
			enriched.MemberName = &name
			enriched.DepartmentID = member.DepartmentID
			enriched.DepartmentName = member.DepartmentName
		}
	}
	return enriched
}

func (s *service) enrichPlatformKeyResponse(ctx context.Context, key types.PlatformKey) (types.PlatformKey, error) {
	lookups, err := s.loadPlatformKeyLookups(ctx)
	if err != nil {
		return types.PlatformKey{}, err
	}
	return enrichPlatformKey(key, lookups), nil
}

func groupDeptIDsFromLookups(lookups platformKeyLookups) map[string][]string {
	groupDeptIDs := make(map[string][]string, len(lookups.groupByID))
	for id, group := range lookups.groupByID {
		groupDeptIDs[id] = append([]string{}, group.DepartmentIDs...)
	}
	return groupDeptIDs
}
