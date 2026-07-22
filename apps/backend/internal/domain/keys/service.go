package keys

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	ListProviderKeys(ctx context.Context) ([]types.ProviderKey, error)
	CreateProviderKey(ctx context.Context, input types.CreateProviderKeyInput) (types.ProviderKey, error)
	CreateProviderKeyForPlatform(ctx context.Context, input types.CreateProviderKeyInput) (types.ProviderKey, error)
	ToggleProviderKey(ctx context.Context, id uuid.UUID, enabled bool) error
	RotateProviderKey(ctx context.Context, id uuid.UUID, newKey string) (types.ProviderKey, error)
	DeleteProviderKey(ctx context.Context, id uuid.UUID) error
	ListPlatformKeys(ctx context.Context, filter types.PlatformKeyListFilter) (types.PageResult[types.PlatformKey], error)
	CreatePlatformKey(ctx context.Context, input types.CreatePlatformKeyInput) (types.PlatformKey, error)
	UpdatePlatformKey(ctx context.Context, id uuid.UUID, input types.UpdatePlatformKeyInput) (types.PlatformKey, error)
	TogglePlatformKey(ctx context.Context, id uuid.UUID, enabled bool) (types.PlatformKey, error)
	RotatePlatformKey(ctx context.Context, id uuid.UUID) (types.PlatformKey, error)
	RevokePlatformKey(ctx context.Context, id uuid.UUID) error
	DeletePlatformKey(ctx context.Context, id uuid.UUID) error
}

// Store is the narrow store surface the keys domain needs.
type Store interface {
	Keys() store.KeysRepository
	BudgetConsumed() store.BudgetConsumedRepository
	Org() store.OrgRepository
	Budget() store.BudgetRepository
	Models() store.ModelsRepository
	PlatformKeyMappings() store.PlatformKeyMappingRepository
	Company() store.CompanyRepository
	CombinedKeySummaries() store.CombinedKeySummaryRepository
	WithTx(ctx context.Context, fn func(store.Store) error) error
}

type service struct {
	cfg              config.Config
	store            Store
	delayer          common.Delayer
	newAPISync       newapisync.KeysNewAPISync
	cacheInvalidator types.PrecheckCacheInvalidator
}

func NewService(cfg config.Config, st Store, newAPISync newapisync.KeysNewAPISync, delayer common.Delayer, opts ...ServiceOption) Service {
	s := &service{
		cfg:              cfg,
		store:            st,
		delayer:          delayer,
		newAPISync:       newAPISync,
		cacheInvalidator: types.NoopPrecheckCacheInvalidator{},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// ServiceOption configures optional dependencies for keys.Service.
type ServiceOption func(*service)

// WithCacheInvalidator sets the gateway precheck cache invalidator.
func WithCacheInvalidator(inv types.PrecheckCacheInvalidator) ServiceOption {
	return func(s *service) {
		if inv != nil {
			s.cacheInvalidator = inv
		}
	}
}

func (s *service) ListProviderKeys(ctx context.Context) ([]types.ProviderKey, error) {
	return s.store.Keys().ProviderKeys(ctx)
}
