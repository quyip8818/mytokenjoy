package remote

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/org/core"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/datasource"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *Service) applySyncDiff(ctx context.Context, platform types.Platform, diff pkgorg.SyncDiff) (types.ImportResult, error) {
	remoteDepts := append([]datasource.RemoteDepartment{}, diff.AddDepartments...)
	remoteDepts = append(remoteDepts, diff.UpdateDepartments...)
	remoteMembers := append([]datasource.RemoteMember{}, diff.AddMembers...)
	remoteMembers = append(remoteMembers, diff.UpdateMembers...)

	result := types.ImportResult{}
	if len(remoteDepts) > 0 || len(remoteMembers) > 0 {
		importResult, err := s.importRemoteSnapshot(ctx, platform, remoteDepts, remoteMembers)
		if err != nil {
			return result, err
		}
		result.SuccessDepartments += importResult.SuccessDepartments
		result.SuccessMembers += importResult.SuccessMembers
	}

	if len(diff.RemoveMembers) == 0 && len(diff.RemoveDepartments) == 0 {
		return result, nil
	}

	err := s.d.Store.WithTx(ctx, func(st store.Store) error {
		members, err := st.Org().Members(ctx)
		if err != nil {
			return err
		}
		membersDeactivated := false
		for _, removed := range diff.RemoveMembers {
			for i := range members {
				if members[i].ID != removed.ID {
					continue
				}
				members[i].Status = types.MemberStatusInactive
				result.SuccessMembers++
				membersDeactivated = true
			}
		}

		nodes, err := st.Org().Nodes().Tree(ctx)
		if err != nil {
			return err
		}
		state, err := core.LoadProvisionState(ctx, st, nodes)
		if err != nil {
			return err
		}
		for _, removed := range diff.RemoveDepartments {
			if err := core.DeprovisionDepartment(state, removed.ID); err != nil {
				return err
			}
			result.SuccessDepartments++
		}

		state.Nodes = core.RecalcDepartmentMemberCounts(state.Nodes, members)
		if err := core.PersistProvisionState(ctx, st, state); err != nil {
			return err
		}
		if err := st.Org().SetMembers(ctx, members); err != nil {
			return err
		}
		if membersDeactivated {
			return core.BumpAuthzRevisionStore(ctx, st)
		}
		return nil
	})
	return result, err
}

func (s *Service) importRemoteSnapshot(
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
	// Root department maps to the local company root
	if externalID == "" || externalID == "0" {
		return core.RootDepartmentID
	}
	if localID, ok := externalToLocal[externalID]; ok {
		return localID
	}
	for _, dept := range pkgorg.FlattenDepartmentTree(departments) {
		if dept.ExternalID != nil && *dept.ExternalID == externalID {
			return dept.ID
		}
	}
	return pkgorg.LocalDeptID(platform, externalID)
}

func resolveParentLocalID(
	parentExternal string,
	platform types.Platform,
	externalToLocal map[string]string,
) string {
	if parentExternal == "" || parentExternal == "0" {
		return core.RootDepartmentID
	}
	if localID, ok := externalToLocal[parentExternal]; ok {
		return localID
	}
	return pkgorg.LocalDeptID(platform, parentExternal)
}

func stringPtr(value string) *string {
	return &value
}
