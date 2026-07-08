package remote

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/org/core"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/integration/datasource"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

type importOptions struct {
	retryExternalIDs map[string]struct{}
}

func (s *Service) ImportDataSource(ctx context.Context) (types.ImportResult, error) {
	provider, platform, err := s.providerForStored(ctx)
	if err != nil {
		return types.ImportResult{}, err
	}
	return s.importFromProvider(ctx, provider, platform, importOptions{})
}

func (s *Service) RetryImport(ctx context.Context, ids []string) (types.ImportResult, error) {
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

func (s *Service) importFromProvider(
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
		importFailures, err := s.d.Store.Org().ImportFailures(ctx)
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

	err = s.d.Store.WithTx(ctx, func(st store.Store) error {
		membersAdded := false
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
		state, err := core.LoadProvisionState(ctx, st, nodes)
		if err != nil {
			return err
		}
		departments := core.DepartmentsFromState(state)

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
				localID = pkgorg.LocalDeptID(platform, remote.ExternalID)
			}

			existing := pkgorg.FindDepartment(departments, localID)
			if existing != nil && pkgorg.IsManualDeptSource(existing.Source) {
				continue
			}

			if existing == nil {
				if err := core.ProvisionDepartment(state, core.ProvisionInput{
					ID: localID, Name: remote.Name, ParentID: parentLocalID, Period: s.d.BudgetPeriod(),
				}); err != nil {
					return err
				}
				node := pkgorg.FindOrgNode(state.Nodes, localID)
				if node != nil {
					node.ExternalID = stringPtr(remote.ExternalID)
					node.Source = stringPtr(types.DeptSourceImported)
					if remote.LeaderUserID != "" {
						node.ManagerID = stringPtr(pkgorg.LocalMemberID(platform, remote.LeaderUserID))
					}
				}
				departments = core.DepartmentsFromState(state)
				externalToLocal[remote.ExternalID] = localID
				changedDeptIDs = append(changedDeptIDs, localID)
				result.SuccessDepartments++
				continue
			}

			if existing.Name != remote.Name {
				if err := core.RenameDepartment(state, localID, remote.Name); err != nil {
					return err
				}
				changedDeptIDs = append(changedDeptIDs, localID)
			}
			node := pkgorg.FindOrgNode(state.Nodes, localID)
			if node != nil {
				node.ExternalID = stringPtr(remote.ExternalID)
				node.Source = stringPtr(types.DeptSourceImported)
				if remote.LeaderUserID != "" {
					node.ManagerID = stringPtr(pkgorg.LocalMemberID(platform, remote.LeaderUserID))
				}
			}
			departments = core.DepartmentsFromState(state)
			result.SuccessDepartments++
		}

		deptNameByID := flattenDeptNames(departments)
		memberIndex := buildMemberExternalIndex(members)

		for _, remote := range remoteMembers {
			localDept := resolveLocalDeptID(departments, platform, remote.DepartmentExternalID, externalToLocal)
			memberID := pkgorg.LocalMemberID(platform, remote.ExternalID)
			if existing, ok := memberIndex[remote.ExternalID]; ok {
				if pkgorg.IsManualMemberSource(existing.Source) {
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
			membersAdded = true
		}

		state.Nodes = core.RecalcDepartmentMemberCounts(state.Nodes, members)

		if err := core.PersistProvisionState(ctx, st, state); err != nil {
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
		if err := st.Org().SetIntegration(ctx, integration); err != nil {
			return err
		}
		if membersAdded {
			return core.BumpAuthzRevisionStore(ctx, st)
		}
		return nil
	})
	if err != nil {
		return types.ImportResult{}, err
	}

	if err := s.d.Store.Org().SetImportFailures(ctx, result.Failures); err != nil {
		return types.ImportResult{}, fmt.Errorf("persist import failures: %w", err)
	}
	if s.d.ModelLimits != nil && len(changedDeptIDs) > 0 {
		if err := s.d.ModelLimits.EnqueueModelLimitsForDepartments(ctx, changedDeptIDs); err != nil {
			return types.ImportResult{}, fmt.Errorf("enqueue model limits: %w", err)
		}
	}
	return result, nil
}
