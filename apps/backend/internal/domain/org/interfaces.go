package org

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type DataSourceService interface {
	GetDataSourceStatus(ctx context.Context) (types.DataSourceStatus, error)
	TestDataSource(ctx context.Context, cred types.Credential) (types.DataSourceTestResult, error)
	UpdateDataSource(ctx context.Context, cred types.Credential, force bool) error
	SearchDataSource(ctx context.Context, keyword string) (types.DataSourceSearchResult, error)
	ImportDataSource(ctx context.Context) (types.ImportResult, error)
	RetryImport(ctx context.Context, ids []string) (types.ImportResult, error)
	GetFieldMappings(ctx context.Context, platform string) ([]types.FieldMapping, error)
	SaveFieldMappings(ctx context.Context, config types.FieldMappingConfig) error
	TestFieldMapping(ctx context.Context, platform, keyword string) (types.MappingTestResult, error)
}

type SyncService interface {
	GetSyncConfig(ctx context.Context) (types.SyncConfig, error)
	UpdateSyncConfig(ctx context.Context, cfg types.SyncConfig) error
	TriggerSync(ctx context.Context) (types.ImportResult, error)
	RunScheduledSync(ctx context.Context) error
	ListSyncLogs(ctx context.Context, page, pageSize int) (types.PageResult[types.SyncLog], error)
}

type DepartmentService interface {
	GetDepartmentTree(ctx context.Context) ([]types.Department, error)
	CreateDepartment(ctx context.Context, name, parentID string) (types.Department, error)
	UpdateDepartment(ctx context.Context, id, name string) (types.Department, error)
	DeleteDepartment(ctx context.Context, id string) error
}

type MemberService interface {
	ListMembers(ctx context.Context, departmentID, keyword string, directOnly bool, page, pageSize int) (types.MemberPageResult, error)
	CreateMember(ctx context.Context, input types.Member) (types.Member, error)
	UpdateMember(ctx context.Context, id string, input types.Member) (types.Member, error)
	DeleteMembers(ctx context.Context, ids []string, currentMemberID string) error
	UpdateMemberStatus(ctx context.Context, ids []string, status string) error
	TransferMembers(ctx context.Context, ids []string, departmentID string) error
	InviteMember() error
	BatchInvite(ctx context.Context, ids []string) (types.BatchInviteResult, error)
	BatchImport(ctx context.Context, rows []types.BatchImportRow) (types.MemberBatchImportResult, error)
}

type RoleService interface {
	ListRoles(ctx context.Context) ([]types.Role, error)
	CreateRole(ctx context.Context, name string, permissions []string) (types.Role, error)
	UpdateRole(ctx context.Context, id, name string, permissions []string) (types.Role, error)
	DeleteRole(ctx context.Context, id string) error
	ListRoleMembers(ctx context.Context, roleID string) ([]types.Member, error)
	AddRoleMember(ctx context.Context, roleID, memberID string) error
	RemoveRoleMember(ctx context.Context, roleID, memberID string) error
	ListPermissions(ctx context.Context) ([]types.Permission, error)
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
