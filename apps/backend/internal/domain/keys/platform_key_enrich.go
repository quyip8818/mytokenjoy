package keys

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
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

func (l platformKeyLookups) members() []types.Member {
	members := make([]types.Member, 0, len(l.memberByID))
	for _, member := range l.memberByID {
		members = append(members, member)
	}
	return members
}

func (l platformKeyLookups) groups() []types.BudgetGroup {
	groups := make([]types.BudgetGroup, 0, len(l.groupByID))
	for _, group := range l.groupByID {
		groups = append(groups, group)
	}
	return groups
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

func (s *service) enrichPlatformKeyUsed(ctx context.Context, key types.PlatformKey, lookups platformKeyLookups) (types.PlatformKey, error) {
	used, found, err := pkgbudget.PlatformKeyConsumed(ctx, s.store.BudgetSnapshots(), s.store.Org().Nodes(), key, lookups.members(), lookups.groups())
	if err != nil {
		return types.PlatformKey{}, err
	}
	if found {
		key.Used = used
	}
	return key, nil
}

func (s *service) enrichPlatformKeyResponse(ctx context.Context, key types.PlatformKey) (types.PlatformKey, error) {
	lookups, err := s.loadPlatformKeyLookups(ctx)
	if err != nil {
		return types.PlatformKey{}, err
	}
	enriched := enrichPlatformKey(key, lookups)
	return s.enrichPlatformKeyUsed(ctx, enriched, lookups)
}
