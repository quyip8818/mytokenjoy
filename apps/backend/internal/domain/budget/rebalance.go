package budget

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

type Rebalancer interface {
	ProcessAxis(ctx context.Context, axisKind string, axisID uuid.UUID) error
}

// RebalanceStore is the narrow store surface the rebalance processor needs.
type RebalanceStore interface {
	BudgetConsumed() store.BudgetConsumedRepository
	Org() store.OrgRepository
	Budget() store.BudgetRepository
	Keys() store.KeysRepository
	PlatformKeyMappings() store.PlatformKeyMappingRepository
	CombinedKeySummaries() store.CombinedKeySummaryRepository
}

type RebalanceService struct {
	cfg   config.Config
	store RebalanceStore
}

func NewRebalanceService(cfg config.Config, st RebalanceStore) *RebalanceService {
	return &RebalanceService{cfg: cfg, store: st}
}

// rebalanceContext holds preloaded data shared across all mappings in a single ProcessAxis call.
type rebalanceContext struct {
	budgetCtx pkgbudget.BudgetContext
}

func (s *RebalanceService) ProcessAxis(ctx context.Context, axisKind string, axisID uuid.UUID) error {
	var mappings []store.PlatformKeyMapping
	var err error
	switch axisKind {
	case store.RebalanceAxisMember:
		mappings, err = s.store.PlatformKeyMappings().ListMappingsByMemberID(ctx, axisID)
	case store.RebalanceAxisProject:
		mappings, err = s.store.PlatformKeyMappings().ListMappingsByProjectID(ctx, axisID)
	case store.RebalanceAxisCompany:
		mappings, err = s.store.PlatformKeyMappings().ListActiveMappingsByCompany(ctx, axisID)
	default:
		return nil
	}
	if err != nil {
		return err
	}

	// Filter to actionable mappings first.
	active := mappings[:0]
	for _, m := range mappings {
		if m.NewAPIKeyID != nil && m.SyncStatus == store.MappingSyncStatusSynced {
			active = append(active, m)
		}
	}
	if len(active) == 0 {
		return nil
	}

	// Preload shared data once for all mappings.
	rctx, err := s.loadRebalanceContext(ctx)
	if err != nil {
		return err
	}

	for _, mapping := range active {
		if err := s.rebalanceKey(ctx, mapping, rctx); err != nil {
			return err
		}
	}
	return nil
}

func (s *RebalanceService) loadRebalanceContext(ctx context.Context) (*rebalanceContext, error) {
	budgetCtx, err := pkgbudget.LoadBudgetContext(ctx, s.store.BudgetConsumed(), s.store.Org(), s.store.Budget(), s.store.Keys(), s.cfg.Clock())
	if err != nil {
		return nil, err
	}
	return &rebalanceContext{
		budgetCtx: budgetCtx,
	}, nil
}

func (s *RebalanceService) rebalanceKey(ctx context.Context, mapping store.PlatformKeyMapping, rctx *rebalanceContext) error {
	key, ok := rctx.budgetCtx.FindPlatformKey(mapping.PlatformKeyID)
	if !ok || key.Status != "active" {
		return nil
	}
	// Token is unlimited on NewAPI — no remote quota to sync.
	// Only refresh the local combined_key_remain for gateway precheck.
	return RefreshPlatformKeyCombined(ctx, s.store, mapping.PlatformKeyID, s.cfg.Clock(), nil)
}

var _ Rebalancer = (*RebalanceService)(nil)
