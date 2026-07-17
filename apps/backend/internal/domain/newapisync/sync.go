package newapisync

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/newapisync/devapi"
	"github.com/tokenjoy/backend/internal/domain/newapisync/modellimits"
	"github.com/tokenjoy/backend/internal/domain/newapisync/platformkey"
	"github.com/tokenjoy/backend/internal/domain/newapisync/policy"
	"github.com/tokenjoy/backend/internal/domain/newapisync/ports"
	"github.com/tokenjoy/backend/internal/domain/newapisync/provider"
	"github.com/tokenjoy/backend/internal/domain/newapisync/provision"
	"github.com/tokenjoy/backend/internal/domain/newapisync/syncdeps"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type NewAPISync struct {
	deps syncdeps.Deps
}

func New(cfg config.Config, st store.Store, client adminport.Port, channelPolicy policy.ChannelPolicy, enqueuer ports.SyncJobEnqueuer) *NewAPISync {
	if enqueuer == nil {
		enqueuer = ports.NoopSyncJobEnqueuer
	}
	return &NewAPISync{
		deps: syncdeps.Deps{
			Cfg:           cfg,
			Store:         st,
			Client:        client,
			Mappings:      st.PlatformKeyMappings(),
			Enqueuer:      enqueuer,
			ChannelPolicy: channelPolicy,
		},
	}
}

func (l *NewAPISync) Enabled() bool {
	return syncdeps.Enabled(l.deps)
}

// Bootstrap synchronously provisions demo platform keys to NewAPI (local only).
func (l *NewAPISync) Bootstrap(ctx context.Context, companyID uuid.UUID) error {
	return provision.Bootstrap(ctx, l.deps, companyID)
}

func (l *NewAPISync) SyncPlatformKeyCreate(ctx context.Context, key types.PlatformKey, departmentID uuid.UUID) (string, error) {
	return platformkey.SyncPlatformKeyCreate(ctx, l.deps, key, departmentID)
}

func (l *NewAPISync) SyncCreatePlatformKey(ctx context.Context, key types.PlatformKey, departmentID uuid.UUID) error {
	return platformkey.SyncCreatePlatformKey(ctx, l.deps, key, departmentID)
}

func (l *NewAPISync) TrySyncCreate(ctx context.Context, platformKeyID uuid.UUID) (string, error) {
	return platformkey.TrySyncCreate(ctx, l.deps, platformKeyID)
}

func (l *NewAPISync) RollbackFailedCreate(ctx context.Context, platformKeyID uuid.UUID) {
	platformkey.RollbackFailedCreate(ctx, l.deps, platformKeyID)
}

func (l *NewAPISync) SyncUpdatePlatformKey(ctx context.Context, platformKeyID uuid.UUID, targetActive *bool) error {
	return platformkey.SyncUpdatePlatformKey(ctx, l.deps, platformKeyID, targetActive)
}

func (l *NewAPISync) SyncRevokePlatformKey(ctx context.Context, platformKeyID uuid.UUID) error {
	return platformkey.SyncRevokePlatformKey(ctx, l.deps, platformKeyID)
}

func (l *NewAPISync) SyncRotatePlatformKey(ctx context.Context, platformKeyID uuid.UUID) (string, error) {
	return platformkey.SyncRotatePlatformKey(ctx, l.deps, platformKeyID)
}

func (l *NewAPISync) DisablePlatformKey(ctx context.Context, platformKeyID uuid.UUID) error {
	return platformkey.DisablePlatformKey(ctx, l.deps, platformKeyID)
}

func (l *NewAPISync) ResolvePlatformKeyBearer(ctx context.Context, platformKeyID uuid.UUID) (string, error) {
	return platformkey.ResolvePlatformKeyBearer(ctx, l.deps, platformKeyID)
}

func (l *NewAPISync) UnreadyPlatformKeyIDs(ctx context.Context) ([]uuid.UUID, error) {
	return provision.UnreadyPlatformKeyIDs(ctx, l.deps)
}

func (l *NewAPISync) EnqueueUpsertProviderKey(ctx context.Context, providerKeyID uuid.UUID) error {
	return provider.EnqueueUpsertProviderKey(ctx, l.deps, providerKeyID)
}

func (l *NewAPISync) SyncUpsertProviderKey(ctx context.Context, providerKeyID uuid.UUID) error {
	return provider.SyncUpsertProviderKey(ctx, l.deps, providerKeyID)
}

func (l *NewAPISync) EnqueueModelLimitsForDepartment(ctx context.Context, departmentID uuid.UUID) error {
	return modellimits.EnqueueModelLimitsForDepartment(ctx, l.deps, departmentID)
}

func (l *NewAPISync) EnqueueModelLimitsForDepartments(ctx context.Context, departmentIDs []uuid.UUID) error {
	return modellimits.EnqueueModelLimitsForDepartments(ctx, l.deps, departmentIDs)
}

func (l *NewAPISync) SyncModelLimitsForDepartment(ctx context.Context, departmentID uuid.UUID) error {
	return modellimits.SyncModelLimitsForDepartment(ctx, l.deps, departmentID)
}

func (l *NewAPISync) EnqueueRebalanceAxis(ctx context.Context, axisKind, axisID string) error {
	if !l.Enabled() {
		return nil
	}
	return l.deps.Enqueuer.InsertRebalance(ctx, company.CompanyID(ctx), axisKind, axisID)
}

var (
	_ Lifecycle               = (*NewAPISync)(nil)
	_ ModelLimitsLifecycle    = (*NewAPISync)(nil)
	_ OverrunKeyControl       = (*NewAPISync)(nil)
	_ KeysNewAPISync          = (*NewAPISync)(nil)
	_ OutboxHandler           = (*NewAPISync)(nil)
	_ devapi.BearerResolver   = (*NewAPISync)(nil)
	_ devapi.ReadinessChecker = (*NewAPISync)(nil)
)
