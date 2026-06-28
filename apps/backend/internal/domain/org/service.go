package org

import (
	"context"
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/datasource"
	"github.com/tokenjoy/backend/internal/notification"
	"github.com/tokenjoy/backend/internal/pkg"
	"github.com/tokenjoy/backend/internal/pkg/roleutil"
	"github.com/tokenjoy/backend/internal/pkg/simulate"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	GetDataSourceStatus() DataSourceStatus
	TestDataSource(ctx context.Context, cred Credential) (DataSourceTestResult, error)
	UpdateDataSource(ctx context.Context, cred Credential, force bool) error
	SearchDataSource(ctx context.Context, keyword string) (DataSourceSearchResult, error)
	ImportDataSource(ctx context.Context) (ImportResult, error)
	RetryImport(ctx context.Context, ids []string) (ImportResult, error)
	GetSyncConfig() SyncConfig
	UpdateSyncConfig(cfg SyncConfig)
	TriggerSync(ctx context.Context) (ImportResult, error)
	RunScheduledSync(ctx context.Context) error
	ListSyncLogs(page, pageSize int) types.PageResult[SyncLog]
	GetDepartmentTree() []Department
	CreateDepartment(ctx context.Context, name, parentID string) (Department, error)
	UpdateDepartment(ctx context.Context, id, name string) (Department, error)
	DeleteDepartment(ctx context.Context, id string) error
	ListMembers(departmentID, keyword string, directOnly bool, page, pageSize int) types.PageResult[Member]
	CreateMember(input Member) (Member, error)
	UpdateMember(id string, input Member) (Member, error)
	DeleteMembers(ctx context.Context, ids []string) error
	UpdateMemberStatus(ctx context.Context, ids []string, status string) error
	TransferMembers(ctx context.Context, ids []string, departmentID string) error
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
	cfg       config.Config
	store     store.Store
	factory   datasource.Factory
	lifecycle relay.Lifecycle
	notifier  notification.Notifier
	delayer   simulate.Delayer
	cryptoKey []byte
	logger    *slog.Logger
}

func NewService(
	cfg config.Config,
	st store.Store,
	factory datasource.Factory,
	lifecycle relay.Lifecycle,
	notifier notification.Notifier,
	logger *slog.Logger,
) Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &service{
		cfg:       cfg,
		store:     st,
		factory:   factory,
		lifecycle: lifecycle,
		notifier:  notifier,
		delayer:   simulate.NewDelayer(cfg.SimulateDelay),
		logger:    logger,
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

func (s *service) GetSyncConfig() SyncConfig {
	return s.store.Org().SyncConfig()
}

func (s *service) UpdateSyncConfig(cfg SyncConfig) {
	_ = s.store.Org().SetSyncConfig(cfg)
}

func (s *service) ListSyncLogs(page, pageSize int) types.PageResult[SyncLog] {
	logs := s.store.Org().SyncLogs()
	items, total, safePage, safeSize := pkg.Paginate(logs, page, pageSize)
	return types.PageResult[SyncLog]{
		Items: items, Total: total, Page: safePage, PageSize: safeSize,
	}
}
