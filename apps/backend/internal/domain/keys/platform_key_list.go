package keys

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
)

func (s *service) ListPlatformKeys(
	ctx context.Context,
	filter types.PlatformKeyListFilter,
) (types.PageResult[types.PlatformKey], error) {
	items, err := pkgbudget.LoadPlatformKeysWithUsed(ctx, s.store.BudgetConsumed(), s.store.Org(), s.store.Budget(), s.store.Keys(), s.cfg.Clock())
	if err != nil {
		return types.PageResult[types.PlatformKey]{}, err
	}

	lookups, err := s.loadPlatformKeyLookups(ctx)
	if err != nil {
		return types.PageResult[types.PlatformKey]{}, err
	}

	var allowedDeptIDs map[string]struct{}

	if filter.DepartmentID != "" {
		departments, err := common.LoadDepartments(ctx, s.store.Org().Nodes())
		if err != nil {
			return types.PageResult[types.PlatformKey]{}, err
		}
		allowedDeptIDs = make(map[string]struct{})
		for _, id := range pkgorg.CollectDescendantDeptIDs(departments, filter.DepartmentID) {
			allowedDeptIDs[id] = struct{}{}
		}
	}

	filtered := make([]types.PlatformKey, 0, len(items))
	for _, key := range items {
		enriched := enrichPlatformKey(key, lookups)
		if !matchesPlatformKeyFilter(enriched, filter, allowedDeptIDs, lookups.projectByID) {
			continue
		}
		filtered = append(filtered, enriched)
	}

	return types.PageResult[types.PlatformKey]{
		Items: filtered, Total: len(filtered), Page: 1, PageSize: 20,
	}, nil
}

func matchesPlatformKeyFilter(
	key types.PlatformKey,
	filter types.PlatformKeyListFilter,
	allowedDeptIDs map[string]struct{},
	projectByID map[string]types.Project,
) bool {
	if filter.MemberID != "" && (key.MemberID == nil || *key.MemberID != filter.MemberID) {
		return false
	}
	if filter.ProjectID != "" && (key.ProjectID == nil || *key.ProjectID != filter.ProjectID) {
		return false
	}
	if filter.Scope != "" && key.Scope != filter.Scope {
		return false
	}
	if filter.DepartmentID == "" {
		return true
	}
	if key.Scope == "member" {
		_, ok := allowedDeptIDs[key.DepartmentID]
		return ok
	}
	if key.ProjectID == nil {
		return false
	}
	project, ok := projectByID[*key.ProjectID]
	if !ok {
		return false
	}
	_, allowed := allowedDeptIDs[project.OwnerDepartmentID]
	return allowed
}
