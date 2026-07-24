// Package port defines cross-domain interfaces consumed by multiple domains.
// Implementations live in integration/ or adapter/ — domain packages only import this package.
package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
)

// KeySyncPort is consumed by the keys domain to synchronize platform/provider keys
// with external systems. Implemented by integration/newapisync.
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

// OverrunKeyControl is consumed by the budget domain's overrun processor
// to disable keys that exceed their budget. Implemented by integration/newapisync.
type OverrunKeyControl interface {
	Enabled() bool
	DisablePlatformKey(ctx context.Context, platformKeyID uuid.UUID) error
}
