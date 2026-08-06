package keys

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
)

// KeySyncPort is consumed by the keys domain to synchronize platform/provider keys
// with external systems. Implemented by newapisync.
type KeySyncPort interface {
	Enabled() bool
	SyncPlatformKeyCreate(ctx context.Context, key types.PlatformKey, departmentID uuid.UUID) (string, error)
	SyncCreatePlatformKey(ctx context.Context, key types.PlatformKey, departmentID uuid.UUID) error
	TrySyncCreate(ctx context.Context, platformKeyID uuid.UUID) (string, error)
	RollbackFailedCreate(ctx context.Context, platformKeyID uuid.UUID)
	SyncUpdatePlatformKey(ctx context.Context, platformKeyID uuid.UUID, targetActive *bool) error
	SyncRevokePlatformKey(ctx context.Context, platformKeyID uuid.UUID) error
	SyncRotatePlatformKey(ctx context.Context, platformKeyID uuid.UUID) (string, error)
	DisablePlatformKey(ctx context.Context, platformKeyID uuid.UUID) error
	EnqueueUpsertProviderKey(ctx context.Context, providerKeyID uuid.UUID) error
	SyncUpsertProviderKey(ctx context.Context, providerKeyID uuid.UUID) error
}
