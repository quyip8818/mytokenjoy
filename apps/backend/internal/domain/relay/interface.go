package relay

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type ModelLimitsEnqueuer interface {
	EnqueueModelLimitsForDepartments(ctx context.Context, departmentIDs []string) error
}

type Lifecycle interface {
	Enabled() bool
	SyncCreatePlatformKey(ctx context.Context, key types.PlatformKey, departmentID string) error
	TrySyncCreate(ctx context.Context, platformKeyID string) (string, error)
	EnqueueUpdatePlatformKey(ctx context.Context, platformKeyID string) error
	SyncUpdatePlatformKey(ctx context.Context, platformKeyID string) error
	SyncRevokePlatformKey(ctx context.Context, platformKeyID string) error
	DisablePlatformKey(ctx context.Context, platformKeyID string) error
	EnqueueUpsertProviderKey(ctx context.Context, providerKeyID string) error
	SyncUpsertProviderKey(ctx context.Context, providerKeyID string) error
	EnqueueModelLimitsForDepartment(ctx context.Context, departmentID string) error
	EnqueueModelLimitsForDepartments(ctx context.Context, departmentIDs []string) error
	SyncModelLimitsForDepartment(ctx context.Context, departmentID string) error
	EnqueueRebalanceAxis(ctx context.Context, axisKind, axisID string) error
	RollbackFailedCreate(ctx context.Context, platformKeyID string)
}
