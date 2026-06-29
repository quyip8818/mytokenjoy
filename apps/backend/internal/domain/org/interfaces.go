package org

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type DataSourceService interface {
	GetDataSourceStatus() types.DataSourceStatus
	TestDataSource(ctx context.Context, cred types.Credential) (types.DataSourceTestResult, error)
	UpdateDataSource(ctx context.Context, cred types.Credential, force bool) error
	SearchDataSource(ctx context.Context, keyword string) (types.DataSourceSearchResult, error)
	ImportDataSource(ctx context.Context) (types.ImportResult, error)
	RetryImport(ctx context.Context, ids []string) (types.ImportResult, error)
}

type SyncService interface {
	GetSyncConfig() types.SyncConfig
	UpdateSyncConfig(cfg types.SyncConfig) error
	TriggerSync(ctx context.Context) (types.ImportResult, error)
	RunScheduledSync(ctx context.Context) error
	ListSyncLogs(page, pageSize int) types.PageResult[types.SyncLog]
}

type DepartmentService interface {
	GetDepartmentTree() []types.Department
	CreateDepartment(ctx context.Context, name, parentID string) (types.Department, error)
	UpdateDepartment(ctx context.Context, id, name string) (types.Department, error)
	DeleteDepartment(ctx context.Context, id string) error
}

type MemberService interface {
	ListMembers(departmentID, keyword string, directOnly bool, page, pageSize int) types.PageResult[types.Member]
	CreateMember(input types.Member) (types.Member, error)
	UpdateMember(id string, input types.Member) (types.Member, error)
	DeleteMembers(ctx context.Context, ids []string) error
	UpdateMemberStatus(ctx context.Context, ids []string, status string) error
	TransferMembers(ctx context.Context, ids []string, departmentID string) error
	InviteMember() error
	BatchInvite(ids []string) types.BatchInviteResult
	BatchImport(rows []types.BatchImportRow) types.MemberBatchImportResult
}

type RoleService interface {
	ListRoles() []types.Role
	CreateRole(name string, permissions []string) (types.Role, error)
	UpdateRole(id, name string, permissions []string) (types.Role, error)
	DeleteRole(id string) error
	ListRoleMembers(roleID string) []types.Member
	AddRoleMember(roleID, memberID string) error
	RemoveRoleMember(roleID, memberID string) error
	ListPermissions() []types.Permission
}

type Service interface {
	DataSourceService
	SyncService
	DepartmentService
	MemberService
	RoleService
}

var (
	_ Service           = (*service)(nil)
	_ DataSourceService = (*service)(nil)
	_ SyncService       = (*service)(nil)
	_ DepartmentService = (*service)(nil)
	_ MemberService     = (*service)(nil)
	_ RoleService       = (*service)(nil)
)
