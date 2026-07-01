package org

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/permission"
	"github.com/tokenjoy/backend/internal/pkg/common"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *service) ListMembers(ctx context.Context, departmentID, keyword string, directOnly bool, page, pageSize int) (types.PageResult[types.Member], error) {
	items, err := s.store.Org().Members(ctx)
	if err != nil {
		return types.PageResult[types.Member]{}, err
	}
	if departmentID != "" {
		departments, err := s.store.Org().Departments(ctx)
		if err != nil {
			return types.PageResult[types.Member]{}, err
		}
		items = pkgorg.FilterMembersByDepartment(items, departments, departmentID, directOnly)
	}
	if keyword != "" {
		filtered := make([]types.Member, 0)
		for _, member := range items {
			if strings.Contains(member.Name, keyword) {
				filtered = append(filtered, member)
			}
		}
		items = filtered
	}
	paged, total, safePage, safeSize := common.Paginate(items, page, pageSize)
	return types.PageResult[types.Member]{
		Items: paged, Total: total, Page: safePage, PageSize: safeSize,
	}, nil
}

func (s *service) CreateMember(ctx context.Context, input types.Member) (types.Member, error) {
	departments, err := s.store.Org().Departments(ctx)
	if err != nil {
		return types.Member{}, err
	}
	dept := pkgorg.FindDepartment(departments, input.DepartmentID)
	if dept == nil {
		return types.Member{}, domain.NewDomainError(domain.StatusNotFound, "types.Department not found")
	}

	deptName := dept.Name
	if path := pkgorg.GetDeptPath(departments, input.DepartmentID); path != nil {
		deptName = *path
	}

	member := types.Member{
		ID:   fmt.Sprintf("m-%d", time.Now().UnixMilli()),
		Name: input.Name, Phone: input.Phone, Email: input.Email,
		DepartmentID: input.DepartmentID, DepartmentName: deptName,
		Status: "active", Roles: []string{permission.RoleMember}, Source: "manual",
	}

	members, err := s.store.Org().Members(ctx)
	if err != nil {
		return types.Member{}, err
	}
	members = append(members, member)
	departments = RecalcDepartmentMemberCounts(departments, members)
	if err := s.store.Org().SetMembers(ctx, members); err != nil {
		return types.Member{}, err
	}
	if err := s.store.Org().SetDepartments(ctx, departments); err != nil {
		return types.Member{}, err
	}
	return member, nil
}

func (s *service) UpdateMember(ctx context.Context, id string, input types.Member) (types.Member, error) {
	members, err := s.store.Org().Members(ctx)
	if err != nil {
		return types.Member{}, err
	}
	for i := range members {
		if members[i].ID == id {
			updated := input
			updated.ID = id
			members[i] = updated
			if err := s.store.Org().SetMembers(ctx, members); err != nil {
				return types.Member{}, err
			}
			return updated, nil
		}
	}
	return types.Member{}, domain.NewDomainError(404, "types.Member not found")
}

func (s *service) DeleteMembers(ctx context.Context, ids []string) error {
	return s.UpdateMemberStatus(ctx, ids, "inactive")
}

func (s *service) UpdateMemberStatus(ctx context.Context, ids []string, status string) error {
	return s.store.WithTx(ctx, func(st store.Store) error {
		members, err := st.Org().Members(ctx)
		if err != nil {
			return err
		}
		keys, err := st.Keys().PlatformKeys(ctx)
		if err != nil {
			return err
		}
		for _, id := range ids {
			for i := range members {
				if members[i].ID == id {
					members[i].Status = status
					if status == "inactive" {
						for j := range keys {
							if keys[j].MemberID != nil && *keys[j].MemberID == id {
								keys[j].Status = "disabled"
							}
						}
					}
				}
			}
		}
		if err := st.Org().SetMembers(ctx, members); err != nil {
			return err
		}
		return st.Keys().SetPlatformKeys(ctx, keys)
	})
}

func (s *service) TransferMembers(ctx context.Context, ids []string, departmentID string) error {
	if len(ids) == 0 {
		return nil
	}

	return s.store.WithTx(ctx, func(st store.Store) error {
		departments, err := st.Org().Departments(ctx)
		if err != nil {
			return err
		}
		target := pkgorg.FindDepartment(departments, departmentID)
		if target == nil {
			return domain.NewDomainError(domain.StatusNotFound, "types.Department not found")
		}

		deptName := target.Name
		if path := pkgorg.GetDeptPath(departments, departmentID); path != nil {
			deptName = *path
		}

		members, err := st.Org().Members(ctx)
		if err != nil {
			return err
		}
		idSet := make(map[string]struct{}, len(ids))
		for _, id := range ids {
			idSet[id] = struct{}{}
		}

		for i := range members {
			if _, ok := idSet[members[i].ID]; !ok {
				continue
			}
			members[i].DepartmentID = departmentID
			members[i].DepartmentName = deptName

			mappings, err := st.Relay().ListMappingsByMemberID(ctx, members[i].ID)
			if err != nil {
				return err
			}
			for _, mapping := range mappings {
				mapping.DepartmentID = departmentID
				if err := st.Relay().UpsertMapping(ctx, mapping); err != nil {
					return err
				}
			}
		}

		departments = RecalcDepartmentMemberCounts(departments, members)
		if err := st.Org().SetMembers(ctx, members); err != nil {
			return err
		}
		return st.Org().SetDepartments(ctx, departments)
	})
}

func (s *service) InviteMember() error {
	return domain.NewDomainError(domain.StatusNotImplemented, "Invite member is not implemented")
}

func (s *service) BatchInvite(ctx context.Context, ids []string) (types.BatchInviteResult, error) {
	members, err := s.store.Org().Members(ctx)
	if err != nil {
		return types.BatchInviteResult{}, err
	}
	targets := make([]types.Member, 0)
	if len(ids) > 0 {
		idSet := make(map[string]struct{}, len(ids))
		for _, id := range ids {
			idSet[id] = struct{}{}
		}
		for _, member := range members {
			if _, ok := idSet[member.ID]; ok {
				targets = append(targets, member)
			}
		}
	} else {
		for _, member := range members {
			if member.Status == "pending" || member.Status == "inactive" {
				targets = append(targets, member)
			}
		}
	}
	return types.BatchInviteResult{Sent: len(targets)}, nil
}

func (s *service) BatchImport(ctx context.Context, rows []types.BatchImportRow) (types.MemberBatchImportResult, error) {
	members, err := s.store.Org().Members(ctx)
	if err != nil {
		return types.MemberBatchImportResult{}, err
	}
	departments, err := s.store.Org().Departments(ctx)
	if err != nil {
		return types.MemberBatchImportResult{}, err
	}
	flat := pkgorg.FlattenDepartmentTree(departments)
	failures := make([]types.MemberBatchImportFailure, 0)
	imported := 0

	for index, row := range rows {
		var dept *types.Department
		for i := range flat {
			if flat[i].Name == row.DepartmentName {
				dept = &flat[i]
				break
			}
		}
		if dept == nil {
			failures = append(failures, types.MemberBatchImportFailure{
				Row: index + 1, Reason: "types.Department not found",
			})
			continue
		}
		members = append(members, types.Member{
			ID:   fmt.Sprintf("m-import-%d-%d", time.Now().UnixMilli(), index),
			Name: row.Name, Phone: row.Phone, Email: row.Email,
			DepartmentID: dept.ID, DepartmentName: dept.Name,
			Status: "active", Roles: []string{permission.RoleMember}, Source: "imported",
		})
		imported++
	}

	if err := s.store.Org().SetMembers(ctx, members); err != nil {
		return types.MemberBatchImportResult{Imported: imported, Failures: append(failures, types.MemberBatchImportFailure{
			Row: 0, Reason: "Failed to persist imported members",
		})}, nil
	}

	return types.MemberBatchImportResult{Imported: imported, Failures: failures}, nil
}
