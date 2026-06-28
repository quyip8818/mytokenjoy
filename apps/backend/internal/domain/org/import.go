package org

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/datasource"
	"github.com/tokenjoy/backend/internal/permission"
	"github.com/tokenjoy/backend/internal/pkg/orgutil"
	"github.com/tokenjoy/backend/internal/store"
)

func localDeptID(platform types.Platform, externalID string) string {
	return fmt.Sprintf("dept-%s-%s", platform, externalID)
}

func localMemberID(platform types.Platform, externalID string) string {
	return fmt.Sprintf("m-%s-%s", platform, externalID)
}

func stringPtr(value string) *string {
	return &value
}

func isManualDeptSource(source *string) bool {
	return source != nil && *source == types.DeptSourceManual
}

func isManualMemberSource(source string) bool {
	return source == types.MemberSourceManual
}

type importOptions struct {
	retryExternalIDs map[string]struct{}
}

func (s *service) ImportDataSource(ctx context.Context) (ImportResult, error) {
	provider, platform, err := s.providerForStored()
	if err != nil {
		return ImportResult{}, err
	}
	return s.importFromProvider(ctx, provider, platform, importOptions{})
}

func (s *service) RetryImport(ctx context.Context, ids []string) (ImportResult, error) {
	provider, platform, err := s.providerForStored()
	if err != nil {
		return ImportResult{}, err
	}
	retrySet := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		retrySet[id] = struct{}{}
	}
	return s.importFromProvider(ctx, provider, platform, importOptions{retryExternalIDs: retrySet})
}

func (s *service) importFromProvider(
	ctx context.Context,
	provider datasource.Provider,
	platform types.Platform,
	opts importOptions,
) (ImportResult, error) {
	retryOnly := len(opts.retryExternalIDs) > 0

	remoteDepts, err := provider.ListDepartments(ctx)
	if err != nil {
		return ImportResult{}, domain.NewDomainError(domain.StatusUnprocessable, err.Error())
	}
	remoteMembers, fetchFailures, err := provider.ListMembers(ctx)
	if err != nil {
		return ImportResult{}, domain.NewDomainError(domain.StatusUnprocessable, err.Error())
	}

	if retryOnly {
		remoteMembers = filterRemoteMembersForRetry(remoteMembers, opts.retryExternalIDs, s.store.Org().ImportFailures())
		if len(remoteMembers) == 0 {
			return ImportResult{
				SuccessMembers: 0, SuccessDepartments: 0,
				Failures: s.store.Org().ImportFailures(),
			}, nil
		}
		deptSet := make(map[string]struct{})
		for _, member := range remoteMembers {
			deptSet[member.DepartmentExternalID] = struct{}{}
		}
		filteredDepts := make([]datasource.RemoteDepartment, 0)
		for _, dept := range remoteDepts {
			if _, ok := deptSet[dept.ExternalID]; ok {
				filteredDepts = append(filteredDepts, dept)
			}
		}
		remoteDepts = filteredDepts
	}

	result := ImportResult{Failures: append([]ImportFailure{}, fetchFailures...)}
	changedDeptIDs := make([]string, 0)

	err = s.store.WithTx(ctx, func(st store.Store) error {
		departments := st.Org().Departments()
		members := st.Org().Members()
		roles := st.Org().Roles()
		state := &ProvisionState{
			Departments: departments,
			BudgetTree:  st.Budget().Tree(),
			Rules:       st.Models().RoutingRules(),
			Models:      st.Models().Models(),
		}

		externalToLocal := buildDeptExternalMap(state.Departments)
		parentMap := make(map[string]string, len(remoteDepts))
		for _, remote := range remoteDepts {
			parentMap[remote.ExternalID] = remote.ParentExternalID
		}
		sortedDepts := sortRemoteDepartments(remoteDepts, parentMap)

		for _, remote := range sortedDepts {
			parentLocalID := resolveParentLocalID(remote.ParentExternalID, platform, externalToLocal)
			localID, exists := externalToLocal[remote.ExternalID]
			if !exists {
				localID = localDeptID(platform, remote.ExternalID)
			}

			existing := orgutil.FindDepartment(state.Departments, localID)
			if existing != nil && isManualDeptSource(existing.Source) {
				continue
			}

			if existing == nil {
				if err := ProvisionDepartment(state, ProvisionInput{
					ID: localID, Name: remote.Name, ParentID: parentLocalID, Period: s.budgetPeriod(),
				}); err != nil {
					return err
				}
				dept := orgutil.FindDepartment(state.Departments, localID)
				if dept != nil {
					dept.ExternalID = stringPtr(remote.ExternalID)
					dept.Source = stringPtr(types.DeptSourceImported)
					if remote.LeaderUserID != "" {
						dept.ManagerID = stringPtr(localMemberID(platform, remote.LeaderUserID))
					}
				}
				externalToLocal[remote.ExternalID] = localID
				changedDeptIDs = append(changedDeptIDs, localID)
				result.SuccessDepartments++
				continue
			}

			if existing.Name != remote.Name {
				if err := RenameDepartment(state, localID, remote.Name); err != nil {
					return err
				}
				changedDeptIDs = append(changedDeptIDs, localID)
			}
			existing.ExternalID = stringPtr(remote.ExternalID)
			existing.Source = stringPtr(types.DeptSourceImported)
			if remote.LeaderUserID != "" {
				existing.ManagerID = stringPtr(localMemberID(platform, remote.LeaderUserID))
			}
			result.SuccessDepartments++
		}

		deptNameByID := flattenDeptNames(state.Departments)
		memberIndex := buildMemberExternalIndex(members)

		for _, remote := range remoteMembers {
			localDept := resolveLocalDeptID(state.Departments, platform, remote.DepartmentExternalID, externalToLocal)
			memberID := localMemberID(platform, remote.ExternalID)
			if existing, ok := memberIndex[remote.ExternalID]; ok {
				if isManualMemberSource(existing.Source) {
					continue
				}
				for i := range members {
					if members[i].ID != existing.ID {
						continue
					}
					members[i].Name = remote.Name
					members[i].Email = remote.Email
					members[i].Phone = remote.Mobile
					members[i].DepartmentID = localDept
					members[i].DepartmentName = deptNameByID[localDept]
					members[i].ExternalID = stringPtr(remote.ExternalID)
					members[i].Source = types.MemberSourceImported
					if members[i].Status == "" {
						members[i].Status = "active"
					}
					break
				}
				result.SuccessMembers++
				continue
			}

			members = append(members, Member{
				ID:             memberID,
				Name:           remote.Name,
				Phone:          remote.Mobile,
				Email:          remote.Email,
				DepartmentID:   localDept,
				DepartmentName: deptNameByID[localDept],
				Status:         "active",
				Roles:          []string{permission.RoleMember},
				Source:         types.MemberSourceImported,
				ExternalID:     stringPtr(remote.ExternalID),
			})
			memberIndex[remote.ExternalID] = members[len(members)-1]
			result.SuccessMembers++
		}

		state.Departments = RecalcDepartmentMemberCounts(state.Departments, members)
		s.recalcRoleMemberCounts(roles)

		if err := st.Org().SetDepartments(state.Departments); err != nil {
			return err
		}
		if err := st.Budget().SetTree(state.BudgetTree); err != nil {
			return err
		}
		if err := st.Models().SetRoutingRules(state.Rules); err != nil {
			return err
		}
		if err := st.Org().SetMembers(members); err != nil {
			return err
		}
		if err := st.Org().SetRoles(roles); err != nil {
			return err
		}

		now := time.Now().Format("2006-01-02 15:04")
		status := st.Org().DataSourceStatus()
		status.Connected = true
		platformCopy := platform
		status.Platform = &platformCopy
		status.LastImport = &now
		status.LastImportResult = &result
		return st.Org().SetDataSourceStatus(status)
	})
	if err != nil {
		return ImportResult{}, err
	}

	_ = s.store.Org().SetImportFailures(result.Failures)
	if s.lifecycle != nil && len(changedDeptIDs) > 0 {
		_ = s.lifecycle.EnqueueModelLimitsForDepartments(changedDeptIDs)
	}
	return result, nil
}

func buildDeptExternalMap(departments []types.Department) map[string]string {
	result := make(map[string]string)
	for _, dept := range orgutil.FlattenDepartmentTree(departments) {
		if dept.ExternalID != nil {
			result[*dept.ExternalID] = dept.ID
		}
	}
	return result
}

func buildMemberExternalIndex(members []Member) map[string]Member {
	result := make(map[string]Member, len(members))
	for _, member := range members {
		if member.ExternalID != nil {
			result[*member.ExternalID] = member
		}
	}
	return result
}

func flattenDeptNames(departments []types.Department) map[string]string {
	result := make(map[string]string)
	for _, dept := range orgutil.FlattenDepartmentTree(departments) {
		result[dept.ID] = dept.Name
		if path := orgutil.GetDeptPath(departments, dept.ID); path != nil {
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
	failures []ImportFailure,
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
	for _, dept := range orgutil.FlattenDepartmentTree(departments) {
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
