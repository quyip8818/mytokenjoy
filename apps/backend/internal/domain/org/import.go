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
	provider, platform, err := s.providerForStored(ctx)
	if err != nil {
		return types.ImportResult{}, err
	}
	return s.importFromProvider(ctx, provider, platform, importOptions{})
}

func (s *service) RetryImport(ctx context.Context, ids []string) (types.ImportResult, error) {
	provider, platform, err := s.providerForStored(ctx)
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
		importFailures, err := s.store.Org().ImportFailures(ctx)
		if err != nil {
			return types.ImportResult{}, err
		}
		remoteMembers = filterRemoteMembersForRetry(remoteMembers, opts.retryExternalIDs, importFailures)
		if len(remoteMembers) == 0 {
			return types.ImportResult{
				SuccessMembers: 0, SuccessDepartments: 0,
				Failures: importFailures,
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
		nodes, err := st.Org().Nodes().Tree(ctx)
		if err != nil {
			return err
		}
		members, err := st.Org().Members(ctx)
		if err != nil {
			return err
		}
		roles, err := st.Org().Roles(ctx)
		if err != nil {
			return err
		}
		state, err := loadProvisionState(ctx, st, nodes)
		if err != nil {
			return err
		}
		departments := departmentsFromState(state)

		externalToLocal := buildDeptExternalMap(departments)
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

			existing := pkgorg.FindDepartment(departments, localID)
			if existing != nil && isManualDeptSource(existing.Source) {
				continue
			}

			if existing == nil {
				if err := ProvisionDepartment(state, ProvisionInput{
					ID: localID, Name: remote.Name, ParentID: parentLocalID, Period: s.budgetPeriod(),
				}); err != nil {
					return err
				}
				node := pkgorg.FindOrgNode(state.Nodes, localID)
				if node != nil {
					node.ExternalID = stringPtr(remote.ExternalID)
					node.Source = stringPtr(types.DeptSourceImported)
					if remote.LeaderUserID != "" {
						node.ManagerID = stringPtr(localMemberID(platform, remote.LeaderUserID))
					}
				}
				departments = departmentsFromState(state)
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
			node := pkgorg.FindOrgNode(state.Nodes, localID)
			if node != nil {
				node.ExternalID = stringPtr(remote.ExternalID)
				node.Source = stringPtr(types.DeptSourceImported)
				if remote.LeaderUserID != "" {
					node.ManagerID = stringPtr(localMemberID(platform, remote.LeaderUserID))
				}
			}
			departments = departmentsFromState(state)
			result.SuccessDepartments++
		}

		deptNameByID := flattenDeptNames(departments)
		memberIndex := buildMemberExternalIndex(members)

		for _, remote := range remoteMembers {
			localDept := resolveLocalDeptID(departments, platform, remote.DepartmentExternalID, externalToLocal)
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

		state.Nodes = RecalcDepartmentMemberCounts(state.Nodes, members)
		if err := s.recalcRoleMemberCounts(ctx, roles); err != nil {
			return err
		}

		if err := persistProvisionState(ctx, st, state); err != nil {
			return err
		}
		if err := st.Org().SetMembers(ctx, members); err != nil {
			return err
		}
		if err := st.Org().SetRoles(ctx, roles); err != nil {
			return err
		}

		now := time.Now().Format("2006-01-02 15:04")
		integration, err := st.Org().Integration(ctx)
		if err != nil {
			return err
		}
		status := integration.ToDataSourceStatus()
		status.Connected = true
		platformCopy := platform
		status.Platform = &platformCopy
		status.LastImport = &now
		status.LastImportResult = &result
		integration.ApplyDataSourceStatus(status)
		return st.Org().SetIntegration(ctx, integration)
	})
	if err != nil {
		return types.ImportResult{}, err
	}

	if err := s.store.Org().SetImportFailures(ctx, result.Failures); err != nil {
		return types.ImportResult{}, fmt.Errorf("persist import failures: %w", err)
	}
	if s.modelLimits != nil && len(changedDeptIDs) > 0 {
		if err := s.modelLimits.EnqueueModelLimitsForDepartments(ctx, changedDeptIDs); err != nil {
			return types.ImportResult{}, fmt.Errorf("enqueue model limits: %w", err)
		}
	}
	return result, nil
}
