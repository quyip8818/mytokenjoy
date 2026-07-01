package store

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type OrgRepository interface {
	DataSourceStatus(ctx context.Context) (types.DataSourceStatus, error)
	SetDataSourceStatus(ctx context.Context, status types.DataSourceStatus) error
	ImportFailures(ctx context.Context) ([]types.ImportFailure, error)
	SetImportFailures(ctx context.Context, failures []types.ImportFailure) error
	SyncConfig(ctx context.Context) (types.SyncConfig, error)
	SetSyncConfig(ctx context.Context, cfg types.SyncConfig) error
	SyncLogs(ctx context.Context) ([]types.SyncLog, error)
	AppendSyncLog(ctx context.Context, log types.SyncLog) error
	Departments(ctx context.Context) ([]types.Department, error)
	SetDepartments(ctx context.Context, departments []types.Department) error
	Members(ctx context.Context) ([]types.Member, error)
	SetMembers(ctx context.Context, members []types.Member) error
	SetMemberPasswordHash(ctx context.Context, memberID, passwordHash string) error
	Roles(ctx context.Context) ([]types.Role, error)
	SetRoles(ctx context.Context, roles []types.Role) error
	Permissions(ctx context.Context) ([]types.Permission, error)
}
