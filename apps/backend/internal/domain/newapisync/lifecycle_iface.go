package newapisync

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
)

type NewAPIGate interface {
	Enabled() bool
}

type PlatformKeyLifecycle interface {
	SyncPlatformKeyCreate(ctx context.Context, key types.PlatformKey, departmentID uuid.UUID) (string, error)
	SyncCreatePlatformKey(ctx context.Context, key types.PlatformKey, departmentID uuid.UUID) error
	TrySyncCreate(ctx context.Context, platformKeyID uuid.UUID) (string, error)
	RollbackFailedCreate(ctx context.Context, platformKeyID uuid.UUID)
	SyncUpdatePlatformKey(ctx context.Context, platformKeyID uuid.UUID, targetActive *bool) error
	SyncRevokePlatformKey(ctx context.Context, platformKeyID uuid.UUID) error
	SyncRotatePlatformKey(ctx context.Context, platformKeyID uuid.UUID) (string, error)
	DisablePlatformKey(ctx context.Context, platformKeyID uuid.UUID) error
}

type ProviderKeyLifecycle interface {
	EnqueueUpsertProviderKey(ctx context.Context, providerKeyID uuid.UUID) error
	SyncUpsertProviderKey(ctx context.Context, providerKeyID uuid.UUID) error
}

type ModelLimitsLifecycle interface {
	EnqueueModelLimitsForDepartment(ctx context.Context, departmentID uuid.UUID) error
	EnqueueModelLimitsForDepartments(ctx context.Context, departmentIDs []uuid.UUID) error
	SyncModelLimitsForDepartment(ctx context.Context, departmentID uuid.UUID) error
}

type RebalanceEnqueuer interface {
	EnqueueRebalanceAxis(ctx context.Context, axisKind string, axisID uuid.UUID) error
}

type Lifecycle interface {
	NewAPIGate
	PlatformKeyLifecycle
	ProviderKeyLifecycle
	ModelLimitsLifecycle
	RebalanceEnqueuer
}

type KeysNewAPISync interface {
	NewAPIGate
	PlatformKeyLifecycle
	ProviderKeyLifecycle
}

type OutboxHandler interface {
	PlatformKeyLifecycle
	ProviderKeyLifecycle
	ModelLimitsLifecycle
}

type OverrunKeyControl interface {
	NewAPIGate
	DisablePlatformKey(ctx context.Context, platformKeyID uuid.UUID) error
}
