package org

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type DataSourceService interface {
	GetDataSourceStatus() DataSourceStatus
	TestDataSource(ctx context.Context, cred Credential) (DataSourceTestResult, error)
	UpdateDataSource(ctx context.Context, cred Credential, force bool) error
	SearchDataSource(ctx context.Context, keyword string) (DataSourceSearchResult, error)
	ImportDataSource(ctx context.Context) (ImportResult, error)
	RetryImport(ctx context.Context, ids []string) (ImportResult, error)
}

type SyncService interface {
	GetSyncConfig() SyncConfig
	UpdateSyncConfig(cfg SyncConfig)
	TriggerSync(ctx context.Context) (ImportResult, error)
	RunScheduledSync(ctx context.Context) error
	ListSyncLogs(page, pageSize int) types.PageResult[SyncLog]
}

type DepartmentService interface {
	GetDepartmentTree() []Department
	CreateDepartment(ctx context.Context, name, parentID string) (Department, error)
	UpdateDepartment(ctx context.Context, id, name string) (Department, error)
	DeleteDepartment(ctx context.Context, id string) error
}

type MemberService interface {
	ListMembers(departmentID, keyword string, directOnly bool, page, pageSize int) types.PageResult[Member]
	CreateMember(input Member) (Member, error)
	UpdateMember(id string, input Member) (Member, error)
	DeleteMembers(ctx context.Context, ids []string) error
	UpdateMemberStatus(ctx context.Context, ids []string, status string) error
	TransferMembers(ctx context.Context, ids []string, departmentID string) error
	InviteMember() error
	BatchInvite(ids []string) BatchInviteResult
	BatchImport(rows []BatchImportRow) MemberBatchImportResult
}

type RoleService interface {
	ListRoles() []Role
	CreateRole(name string, permissions []string) Role
	UpdateRole(id, name string, permissions []string) (Role, error)
	DeleteRole(id string) error
	ListRoleMembers(roleID string) []Member
	AddRoleMember(roleID, memberID string) error
	RemoveRoleMember(roleID, memberID string) error
	ListPermissions() []Permission
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
