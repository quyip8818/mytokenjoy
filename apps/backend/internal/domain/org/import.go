package org

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/integration/datasource"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

type importOptions struct {
	retryExternalIDs map[string]struct{}
}

func (s *service) ImportDataSource(ctx context.Context) (types.ImportResult, error) {
	provider, platform, err := s.providerForStored()
	if err != nil {
		return types.ImportResult{}, err
	}
	return s.importFromProvider(ctx, provider, platform, importOptions{})
}

func (s *service) RetryImport(ctx context.Context, ids []string) (types.ImportResult, error) {
	provider, platform, err := s.providerForStored()
	if err != nil {
		return types.ImportResult{}, err
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
) (types.ImportResult, error) {
	retryOnly := len(opts.retryExternalIDs) > 0

	remoteDepts, err := provider.ListDepartments(ctx)
	if err != nil {
		return types.ImportResult{}, domain.NewDomainError(domain.StatusUnprocessable, err.Error())
	}
	remoteMembers, fetchFailures, err := provider.ListMembers(ctx)
	if err != nil {
		return types.ImportResult{}, domain.NewDomainError(domain.StatusUnprocessable, err.Error())
	}

	if retryOnly {
		remoteMembers = filterRemoteMembersForRetry(remoteMembers, opts.retryExternalIDs, s.store.Org().ImportFailures())
		if len(remoteMembers) == 0 {
			return types.ImportResult{
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

	result := types.ImportResult{Failures: append([]types.ImportFailure{}, fetchFailures...)}
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

			existing := pkgorg.FindDepartment(state.Departments, localID)
			if existing != nil && isManualDeptSource(existing.Source) {
				continue
			}

			if existing == nil {
				if err := ProvisionDepartment(state, ProvisionInput{
					ID: localID, Name: remote.Name, ParentID: parentLocalID, Period: s.budgetPeriod(),
				}); err != nil {
					return err
				}
				dept := pkgorg.FindDepartment(state.Departments, localID)
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

			members = append(members, types.Member{
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
		return types.ImportResult{}, err
	}

	if err := s.store.Org().SetImportFailures(result.Failures); err != nil {
		return types.ImportResult{}, fmt.Errorf("persist import failures: %w", err)
	}
	if s.lifecycle != nil && len(changedDeptIDs) > 0 {
		if err := s.lifecycle.EnqueueModelLimitsForDepartments(changedDeptIDs); err != nil {
			return types.ImportResult{}, fmt.Errorf("enqueue model limits: %w", err)
		}
	}
	return result, nil
}
