package keys

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
)

type platformKeyLookups struct {
	memberByID  map[string]types.Member
	projectByID map[string]types.Project
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
	projects, err := s.store.Budget().Projects(ctx)
	if err != nil {
		return platformKeyLookups{}, err
	}
	projectByID := make(map[string]types.Project, len(projects))
	for _, project := range projects {
		projectByID[project.ID] = project
	}
	return platformKeyLookups{memberByID: memberByID, projectByID: projectByID}, nil
}

func (l platformKeyLookups) members() []types.Member {
	members := make([]types.Member, 0, len(l.memberByID))
	for _, member := range l.memberByID {
		members = append(members, member)
	}
	return members
}

func (l platformKeyLookups) projects() []types.Project {
	projects := make([]types.Project, 0, len(l.projectByID))
	for _, project := range l.projectByID {
		projects = append(projects, project)
	}
	return projects
}

func enrichPlatformKey(key types.PlatformKey, lookups platformKeyLookups) types.PlatformKey {
	enriched := key
	enriched.MemberName = nil
	enriched.ProjectName = nil

	if key.ProjectID != nil {
		enriched.Scope = "project"
		if project, ok := lookups.projectByID[*key.ProjectID]; ok {
			name := project.Name
			enriched.ProjectName = &name
		}
	} else {
		enriched.Scope = "member"
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
	used, found, err := pkgbudget.PlatformKeyConsumed(ctx, s.store.BudgetConsumed(), s.store.Org().Nodes(), key, lookups.members(), lookups.projects(), s.cfg.Clock())
	if err != nil {
		return types.PlatformKey{}, err
	}
	if found {
		key.Consumed = used
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
