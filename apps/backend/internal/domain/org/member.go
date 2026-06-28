package org

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/permission"
	"github.com/tokenjoy/backend/internal/pkg"
	"github.com/tokenjoy/backend/internal/pkg/orgutil"
	"github.com/tokenjoy/backend/internal/pkg/queryutil"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *service) ListMembers(departmentID, keyword string, directOnly bool, page, pageSize int) types.PageResult[Member] {
	items := s.store.Org().Members()
	if departmentID != "" {
		items = queryutil.FilterMembersByDepartment(items, s.store.Org().Departments(), departmentID, directOnly)
	}
	if keyword != "" {
		filtered := make([]Member, 0)
		for _, member := range items {
			if strings.Contains(member.Name, keyword) {
				filtered = append(filtered, member)
			}
		}
		items = filtered
	}
	paged, total, safePage, safeSize := pkg.Paginate(items, page, pageSize)
	return types.PageResult[Member]{
		Items: paged, Total: total, Page: safePage, PageSize: safeSize,
	}
}

func (s *service) CreateMember(input Member) Member {
	return Member{
		ID:   fmt.Sprintf("m-%d", time.Now().UnixMilli()),
		Name: input.Name, Phone: input.Phone, Email: input.Email,
		DepartmentID: input.DepartmentID, DepartmentName: input.DepartmentName,
		Status: "active", Roles: []string{permission.RoleMember}, Source: "manual",
	}
}

func (s *service) UpdateMember(id string, input Member) (Member, error) {
	members := s.store.Org().Members()
	for i := range members {
		if members[i].ID == id {
			updated := input
			updated.ID = id
			members[i] = updated
			if err := s.store.Org().SetMembers(members); err != nil {
				return Member{}, err
			}
			return updated, nil
		}
	}
	return Member{}, domain.NewDomainError(404, "Member not found")
}

func (s *service) DeleteMembers(ids []string) error {
	_ = ids
	return nil
}

func (s *service) UpdateMemberStatus(ids []string, status string) error {
	return s.store.WithTx(context.Background(), func(st store.Store) error {
		members := st.Org().Members()
		keys := st.Keys().PlatformKeys()
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
		if err := st.Org().SetMembers(members); err != nil {
			return err
		}
		return st.Keys().SetPlatformKeys(keys)
	})
}

func (s *service) TransferMembers(ids []string, departmentID string) error {
	_ = ids
	_ = departmentID
	return nil
}

func (s *service) InviteMember() error {
	return nil
}

func (s *service) BatchInvite(ids []string) BatchInviteResult {
	members := s.store.Org().Members()
	targets := make([]Member, 0)
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
	return BatchInviteResult{Sent: len(targets)}
}

func (s *service) BatchImport(rows []BatchImportRow) MemberBatchImportResult {
	members := s.store.Org().Members()
	flat := orgutil.FlattenDepartmentTree(s.store.Org().Departments())
	failures := make([]MemberBatchImportFailure, 0)
	imported := 0

	for index, row := range rows {
		var dept *Department
		for i := range flat {
			if flat[i].Name == row.DepartmentName {
				dept = &flat[i]
				break
			}
		}
		if dept == nil {
			failures = append(failures, MemberBatchImportFailure{
				Row: index + 1, Reason: "Department not found",
			})
			continue
		}
		members = append(members, Member{
			ID:   fmt.Sprintf("m-import-%d-%d", time.Now().UnixMilli(), index),
			Name: row.Name, Phone: row.Phone, Email: row.Email,
			DepartmentID: dept.ID, DepartmentName: dept.Name,
			Status: "active", Roles: []string{permission.RoleMember}, Source: "imported",
		})
		imported++
	}

	_ = s.store.Org().SetMembers(members)
	return MemberBatchImportResult{Imported: imported, Failures: failures}
}
