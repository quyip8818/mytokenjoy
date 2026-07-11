package newapisync

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
)

type NewAPIGate interface {
	Enabled() bool
}

type PlatformKeyLifecycle interface {
	SyncCreatePlatformKey(ctx context.Context, key types.PlatformKey, departmentID string) error
	TrySyncCreate(ctx context.Context, platformKeyID string) (string, error)
	RollbackFailedCreate(ctx context.Context, platformKeyID string)
	SyncUpdatePlatformKey(ctx context.Context, platformKeyID string, targetActive *bool) error
	SyncRevokePlatformKey(ctx context.Context, platformKeyID string) error
	SyncRotatePlatformKey(ctx context.Context, platformKeyID string) (string, error)
	DisablePlatformKey(ctx context.Context, platformKeyID string) error
}

type ProviderKeyLifecycle interface {
	EnqueueUpsertProviderKey(ctx context.Context, providerKeyID string) error
	SyncUpsertProviderKey(ctx context.Context, providerKeyID string) error
}

type ModelLimitsLifecycle interface {
	EnqueueModelLimitsForDepartment(ctx context.Context, departmentID string) error
	EnqueueModelLimitsForDepartments(ctx context.Context, departmentIDs []string) error
	SyncModelLimitsForDepartment(ctx context.Context, departmentID string) error
}

type RebalanceEnqueuer interface {
	EnqueueRebalanceAxis(ctx context.Context, axisKind, axisID string) error
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
	DisablePlatformKey(ctx context.Context, platformKeyID string) error
}
