package relay

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type Lifecycle interface {
	Enabled() bool
	SyncCreatePlatformKey(ctx context.Context, key types.PlatformKey, departmentID string) error
	TrySyncCreate(ctx context.Context, platformKeyID string) (string, error)
	EnqueueUpdatePlatformKey(platformKeyID string) error
	SyncUpdatePlatformKey(ctx context.Context, platformKeyID string) error
	SyncRevokePlatformKey(ctx context.Context, platformKeyID string) error
	DisablePlatformKey(ctx context.Context, platformKeyID string) error
	EnqueueUpsertProviderKey(providerKeyID string) error
	SyncUpsertProviderKey(ctx context.Context, providerKeyID string) error
	EnqueueModelLimitsForDepartment(departmentID string) error
	EnqueueModelLimitsForDepartments(departmentIDs []string) error
	SyncModelLimitsForDepartment(ctx context.Context, departmentID string) error
	EnqueueRebalanceAxis(axisKind, axisID string) error
	RollbackFailedCreate(platformKeyID string)
}
