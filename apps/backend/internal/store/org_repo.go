package store

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type OrgRepository interface {
	Integration(ctx context.Context) (types.OrgIntegration, error)
	SetIntegration(ctx context.Context, integration types.OrgIntegration) error
	GetIntegrationCredential(ctx context.Context) (*types.StoredCredential, error)
	SaveIntegrationCredential(ctx context.Context, platform types.Platform, encrypted []byte) error
	ClearIntegrationCredential(ctx context.Context) error
	ImportFailures(ctx context.Context) ([]types.ImportFailure, error)
	SetImportFailures(ctx context.Context, failures []types.ImportFailure) error
	SyncLogs(ctx context.Context) ([]types.SyncLog, error)
	AppendSyncLog(ctx context.Context, log types.SyncLog) error
	Nodes() OrgNodeRepository
	Members(ctx context.Context) ([]types.Member, error)
	MemberByID(ctx context.Context, memberID string) (*types.Member, error)
	MemberPersonalQuota(ctx context.Context, memberID string) (float64, bool, error)
	SetMembers(ctx context.Context, members []types.Member) error
	UpdateMemberPersonalQuota(ctx context.Context, memberID string, personalQuota float64) error
	SetMemberPasswordHash(ctx context.Context, memberID, passwordHash string) error
	Roles(ctx context.Context) ([]types.Role, error)
	SetRoles(ctx context.Context, roles []types.Role) error
	Permissions(ctx context.Context) ([]types.Permission, error)
}
