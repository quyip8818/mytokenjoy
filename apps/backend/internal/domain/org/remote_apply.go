package org

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/datasource"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *service) applySyncDiff(ctx context.Context, platform types.Platform, diff syncDiff) (types.ImportResult, error) {
	remoteDepts := append([]datasource.RemoteDepartment{}, diff.addDepartments...)
	remoteDepts = append(remoteDepts, diff.updateDepartments...)
	remoteMembers := append([]datasource.RemoteMember{}, diff.addMembers...)
	remoteMembers = append(remoteMembers, diff.updateMembers...)

	result := types.ImportResult{}
	if len(remoteDepts) > 0 || len(remoteMembers) > 0 {
		importResult, err := s.importRemoteSnapshot(ctx, platform, remoteDepts, remoteMembers)
		if err != nil {
			return result, err
		}
		result.SuccessDepartments += importResult.SuccessDepartments
		result.SuccessMembers += importResult.SuccessMembers
	}

	if len(diff.removeMembers) == 0 && len(diff.removeDepartment) == 0 {
		return result, nil
	}

	err := s.store.WithTx(ctx, func(st store.Store) error {
		members, err := st.Org().Members(ctx)
		if err != nil {
			return err
		}
		for _, removed := range diff.removeMembers {
			for i := range members {
				if members[i].ID != removed.ID {
					continue
				}
				members[i].Status = "inactive"
				result.SuccessMembers++
			}
		}

		departments, err := st.Org().Departments(ctx)
		if err != nil {
			return err
		}
		state, err := loadProvisionState(ctx, st, departments)
		if err != nil {
			return err
		}
		for _, removed := range diff.removeDepartment {
			if err := DeprovisionDepartment(state, removed.ID); err != nil {
				return err
			}
			result.SuccessDepartments++
		}

		state.Departments = RecalcDepartmentMemberCounts(state.Departments, members)
		if err := st.Org().SetDepartments(ctx, state.Departments); err != nil {
			return err
		}
		if err := st.Budget().SetTree(ctx, state.BudgetTree); err != nil {
			return err
		}
		if err := st.Models().SetRoutingRules(ctx, state.Rules); err != nil {
			return err
		}
		return st.Org().SetMembers(ctx, members)
	})
	return result, err
}

func (s *service) importRemoteSnapshot(
	ctx context.Context,
	platform types.Platform,
	remoteDepts []datasource.RemoteDepartment,
	remoteMembers []datasource.RemoteMember,
) (types.ImportResult, error) {
	provider := &fixedProvider{departments: remoteDepts, members: remoteMembers}
	return s.importFromProvider(ctx, provider, platform, importOptions{})
}

func buildDeptExternalMap(departments []types.Department) map[string]string {
	result := make(map[string]string)
	for _, dept := range pkgorg.FlattenDepartmentTree(departments) {
		if dept.ExternalID != nil {
			result[*dept.ExternalID] = dept.ID
		}
	}
	return result
}

func buildMemberExternalIndex(members []types.Member) map[string]types.Member {
	result := make(map[string]types.Member, len(members))
	for _, member := range members {
		if member.ExternalID != nil {
			result[*member.ExternalID] = member
		}
	}
	return result
}

func flattenDeptNames(departments []types.Department) map[string]string {
	result := make(map[string]string)
	for _, dept := range pkgorg.FlattenDepartmentTree(departments) {
		result[dept.ID] = dept.Name
		if path := pkgorg.GetDeptPath(departments, dept.ID); path != nil {
			result[dept.ID] = *path
		}
	}
	return result
}

func sortRemoteDepartments(
	remote []datasource.RemoteDepartment,
	_ map[string]string,
) []datasource.RemoteDepartment {
	children := make(map[string][]datasource.RemoteDepartment)
	roots := make([]datasource.RemoteDepartment, 0)
	for _, dept := range remote {
		parentExternal := dept.ParentExternalID
		if parentExternal == "" || parentExternal == "0" {
			roots = append(roots, dept)
			continue
		}
		children[parentExternal] = append(children[parentExternal], dept)
	}
	ordered := make([]datasource.RemoteDepartment, 0, len(remote))
	var walk func(parentExternal string)
	walk = func(parentExternal string) {
		next := children[parentExternal]
		for _, dept := range next {
			ordered = append(ordered, dept)
			walk(dept.ExternalID)
		}
	}
	for _, root := range roots {
		ordered = append(ordered, root)
		walk(root.ExternalID)
	}
	if len(ordered) == 0 {
		return remote
	}
	return ordered
}

func filterRemoteMembersForRetry(
	members []datasource.RemoteMember,
	retryIDs map[string]struct{},
	failures []types.ImportFailure,
) []datasource.RemoteMember {
	targets := make(map[string]struct{})
	for _, failure := range failures {
		if _, ok := retryIDs[failure.ID]; !ok {
			continue
		}
		if failure.EmployeeID != "" {
			targets[failure.EmployeeID] = struct{}{}
		}
		targets[failure.ID] = struct{}{}
	}
	filtered := make([]datasource.RemoteMember, 0)
	for _, member := range members {
		if _, ok := targets[member.ExternalID]; ok {
			filtered = append(filtered, member)
			continue
		}
		if member.EmployeeNo != "" {
			if _, ok := targets[member.EmployeeNo]; ok {
				filtered = append(filtered, member)
			}
		}
	}
	return filtered
}

func resolveLocalDeptID(
	departments []types.Department,
	platform types.Platform,
	externalID string,
	externalToLocal map[string]string,
) string {
	if localID, ok := externalToLocal[externalID]; ok {
		return localID
	}
	for _, dept := range pkgorg.FlattenDepartmentTree(departments) {
		if dept.ExternalID != nil && *dept.ExternalID == externalID {
			return dept.ID
		}
	}
	return localDeptID(platform, externalID)
}

func resolveParentLocalID(
	parentExternal string,
	platform types.Platform,
	externalToLocal map[string]string,
) string {
	if parentExternal == "" || parentExternal == "0" {
		return RootDepartmentID
	}
	if localID, ok := externalToLocal[parentExternal]; ok {
		return localID
	}
	return localDeptID(platform, parentExternal)
}
