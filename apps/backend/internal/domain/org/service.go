package org

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/permission"
	"github.com/tokenjoy/backend/internal/pkg"
	"github.com/tokenjoy/backend/internal/pkg/orgutil"
	"github.com/tokenjoy/backend/internal/pkg/queryutil"
	"github.com/tokenjoy/backend/internal/pkg/roleutil"
	"github.com/tokenjoy/backend/internal/pkg/simulate"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	GetDataSourceStatus() DataSourceStatus
	TestDataSource(ctx context.Context) (DataSourceTestResult, error)
	UpdateDataSource() DataSourceStatus
	SearchDataSource(keyword string) DataSourceSearchResult
	ImportDataSource(ctx context.Context) (ImportResult, error)
	RetryImport(ctx context.Context) (ImportResult, error)
	GetSyncConfig() SyncConfig
	UpdateSyncConfig(cfg SyncConfig)
	TriggerSync(ctx context.Context) (ImportResult, error)
	ListSyncLogs(page, pageSize int) types.PageResult[SyncLog]
	GetDepartmentTree() []Department
	CreateDepartment(name, parentID string) Department
	UpdateDepartment(id, name string) Department
	DeleteDepartment(id string) error
	ListMembers(departmentID, keyword string, directOnly bool, page, pageSize int) types.PageResult[Member]
	CreateMember(input Member) Member
	UpdateMember(id string, input Member) (Member, error)
	DeleteMembers(ids []string) error
	UpdateMemberStatus(ids []string, status string) error
	TransferMembers(ids []string, departmentID string) error
	InviteMember() error
	BatchInvite(ids []string) BatchInviteResult
	BatchImport(rows []BatchImportRow) MemberBatchImportResult
	ListRoles() []Role
	CreateRole(name string, permissions []string) Role
	UpdateRole(id, name string, permissions []string) (Role, error)
	DeleteRole(id string) error
	ListRoleMembers(roleID string) []Member
	AddRoleMember(roleID, memberID string) error
	RemoveRoleMember(roleID, memberID string) error
	ListPermissions() []Permission
}

type service struct {
	cfg     config.Config
	store   store.Store
	delayer simulate.Delayer
}

func NewService(cfg config.Config, st store.Store) Service {
	return &service{
		cfg:     cfg,
		store:   st,
		delayer: simulate.NewDelayer(cfg.SimulateDelay),
	}
}

func (s *service) GetDataSourceStatus() DataSourceStatus {
	return s.store.Org().DataSourceStatus()
}

func (s *service) TestDataSource(ctx context.Context) (DataSourceTestResult, error) {
	if err := s.delayer.Wait(ctx, time.Second); err != nil {
		return DataSourceTestResult{}, err
	}
	return DataSourceTestResult{Success: true}, nil
}

func (s *service) UpdateDataSource() DataSourceStatus {
	platform := PlatformFeishu
	status := s.store.Org().DataSourceStatus()
	status.Connected = true
	status.Platform = &platform
	s.store.Org().SetDataSourceStatus(status)
	return status
}

func (s *service) SearchDataSource(keyword string) DataSourceSearchResult {
	trimmed := strings.TrimSpace(keyword)
	if trimmed == "" {
		return DataSourceSearchResult{}
	}

	members := s.store.Org().Members()
	departments := s.store.Org().Departments()
	for _, member := range members {
		if strings.Contains(member.Name, trimmed) ||
			strings.Contains(member.Phone, trimmed) ||
			strings.Contains(member.Email, trimmed) {
			department := member.DepartmentName
			if path := orgutil.GetDeptPath(departments, member.DepartmentID); path != nil {
				department = *path
			}
			return DataSourceSearchResult{
				Name: member.Name, Department: department, MappingOK: true,
			}
		}
	}
	return DataSourceSearchResult{}
}

func (s *service) ImportDataSource(ctx context.Context) (ImportResult, error) {
	if err := s.delayer.Wait(ctx, 2*time.Second); err != nil {
		return ImportResult{}, err
	}
	return ImportResult{
		SuccessMembers: 120, SuccessDepartments: 5,
		Failures: s.store.Org().ImportFailures(),
	}, nil
}

func (s *service) RetryImport(ctx context.Context) (ImportResult, error) {
	if err := s.delayer.Wait(ctx, 500*time.Millisecond); err != nil {
		return ImportResult{}, err
	}
	return ImportResult{
		SuccessMembers: 1, SuccessDepartments: 0, Failures: []ImportFailure{},
	}, nil
}

func (s *service) GetSyncConfig() SyncConfig {
	return s.store.Org().SyncConfig()
}

func (s *service) UpdateSyncConfig(cfg SyncConfig) {
	s.store.Org().SetSyncConfig(cfg)
}

func (s *service) TriggerSync(ctx context.Context) (ImportResult, error) {
	if err := s.delayer.Wait(ctx, 1500*time.Millisecond); err != nil {
		return ImportResult{}, err
	}
	return ImportResult{
		SuccessMembers: 3, SuccessDepartments: 0, Failures: []ImportFailure{},
	}, nil
}

func (s *service) ListSyncLogs(page, pageSize int) types.PageResult[SyncLog] {
	logs := s.store.Org().SyncLogs()
	items, total, safePage, safeSize := pkg.Paginate(logs, page, pageSize)
	return types.PageResult[SyncLog]{
		Items: items, Total: total, Page: safePage, PageSize: safeSize,
	}
}

func (s *service) GetDepartmentTree() []Department {
	return s.store.Org().Departments()
}

func (s *service) CreateDepartment(name, parentID string) Department {
	return Department{
		ID:   fmt.Sprintf("dept-%d", time.Now().UnixMilli()),
		Name: name, ParentID: &parentID, MemberCount: 0,
	}
}

func (s *service) UpdateDepartment(id, name string) Department {
	return Department{
		ID: id, Name: name, ParentID: nil, MemberCount: 0,
	}
}

func (s *service) DeleteDepartment(id string) error {
	_ = id
	return nil
}

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
			s.store.Org().SetMembers(members)
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
	members := s.store.Org().Members()
	keys := s.store.Keys().PlatformKeys()
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
	s.store.Org().SetMembers(members)
	s.store.Keys().SetPlatformKeys(keys)
	return nil
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

	s.store.Org().SetMembers(members)
	return MemberBatchImportResult{Imported: imported, Failures: failures}
}

func (s *service) ListRoles() []Role {
	return s.store.Org().Roles()
}

func (s *service) CreateRole(name string, permissions []string) Role {
	roles := s.store.Org().Roles()
	role := Role{
		ID:   fmt.Sprintf("role-%d", time.Now().UnixMilli()),
		Name: name, Type: "custom", Permissions: permissions, MemberCount: 0,
	}
	roles = append(roles, role)
	s.store.Org().SetRoles(roles)
	return role
}

func (s *service) UpdateRole(id, name string, permissions []string) (Role, error) {
	roles := s.store.Org().Roles()
	for i := range roles {
		if roles[i].ID == id {
			roles[i].Name = name
			roles[i].Permissions = permissions
			s.store.Org().SetRoles(roles)
			return roles[i], nil
		}
	}
	return Role{}, domain.NewDomainError(404, "Not found")
}

func (s *service) DeleteRole(id string) error {
	roles := s.store.Org().Roles()
	idx := -1
	for i := range roles {
		if roles[i].ID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		return domain.NewDomainError(404, "Not found")
	}
	role := roles[idx]
	if role.Type == "preset" {
		return domain.NewDomainError(400, "Cannot delete preset role")
	}

	members := s.store.Org().Members()
	for i := range members {
		filtered := make([]string, 0, len(members[i].Roles))
		for _, roleName := range members[i].Roles {
			if roleName != role.Name {
				filtered = append(filtered, roleName)
			}
		}
		members[i].Roles = filtered
	}
	s.store.Org().SetMembers(members)

	roles = append(roles[:idx], roles[idx+1:]...)
	s.recalcRoleMemberCounts(roles)
	s.store.Org().SetRoles(roles)
	return nil
}

func (s *service) ListRoleMembers(roleID string) []Member {
	roles := s.store.Org().Roles()
	var role *Role
	for i := range roles {
		if roles[i].ID == roleID {
			role = &roles[i]
			break
		}
	}
	if role == nil {
		return []Member{}
	}

	members := s.store.Org().Members()
	result := make([]Member, 0)
	for _, member := range members {
		for _, roleName := range member.Roles {
			if roleName == role.Name {
				result = append(result, member)
				break
			}
		}
	}
	return result
}

func (s *service) AddRoleMember(roleID, memberID string) error {
	roles := s.store.Org().Roles()
	members := s.store.Org().Members()

	var role *Role
	for i := range roles {
		if roles[i].ID == roleID {
			role = &roles[i]
			break
		}
	}
	if role == nil {
		return nil
	}

	for i := range members {
		if members[i].ID != memberID {
			continue
		}
		if !roleutil.ContainsRole(members[i].Roles, role.Name) {
			members[i].Roles = append(members[i].Roles, role.Name)
			s.recalcRoleMemberCounts(roles)
			s.store.Org().SetMembers(members)
			s.store.Org().SetRoles(roles)
		}
		break
	}
	return nil
}

func (s *service) RemoveRoleMember(roleID, memberID string) error {
	roles := s.store.Org().Roles()
	members := s.store.Org().Members()

	var role *Role
	for i := range roles {
		if roles[i].ID == roleID {
			role = &roles[i]
			break
		}
	}
	var member *Member
	for i := range members {
		if members[i].ID == memberID {
			member = &members[i]
			break
		}
	}
	if role == nil || member == nil {
		return domain.NewDomainError(404, "Not found")
	}
	if role.Name == permission.RoleMember {
		return domain.NewDomainError(400, "Cannot remove base member role")
	}

	filtered := make([]string, 0, len(member.Roles))
	for _, roleName := range member.Roles {
		if roleName != role.Name {
			filtered = append(filtered, roleName)
		}
	}
	for i := range members {
		if members[i].ID == memberID {
			members[i].Roles = filtered
			break
		}
	}
	s.recalcRoleMemberCounts(roles)
	s.store.Org().SetMembers(members)
	s.store.Org().SetRoles(roles)
	return nil
}

func (s *service) ListPermissions() []Permission {
	return s.store.Org().Permissions()
}

func (s *service) recalcRoleMemberCounts(roles []Role) {
	members := s.store.Org().Members()
	for i := range roles {
		roles[i].MemberCount = roleutil.CountMembersByRole(members, roles[i].Name)
	}
}
