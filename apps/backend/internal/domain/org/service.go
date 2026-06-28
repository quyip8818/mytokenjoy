package org

import (
	"context"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
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

func (s *service) recalcRoleMemberCounts(roles []Role) {
	members := s.store.Org().Members()
	for i := range roles {
		roles[i].MemberCount = roleutil.CountMembersByRole(members, roles[i].Name)
	}
}

func (s *service) ListPermissions() []Permission {
	return s.store.Org().Permissions()
}
