package budget

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
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
	cache CombinedKeyCache
}

func NewRebalanceService(cfg config.Config, st RebalanceStore, opts ...RebalanceOption) *RebalanceService {
	s := &RebalanceService{cfg: cfg, store: st}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// RebalanceOption configures optional dependencies for RebalanceService.
type RebalanceOption func(*RebalanceService)

// WithRebalanceCache sets the Redis cache used to propagate combined_key_remain updates.
func WithRebalanceCache(cache CombinedKeyCache) RebalanceOption {
	return func(s *RebalanceService) { s.cache = cache }
}

// ProcessAxis recomputes combined_key_remain for all active platform keys on the given axis.
// It loads budget context once, computes remain for all affected keys in batch, then
// persists and caches the results in a single round-trip each.
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

	// Filter to actionable mappings (synced with a live NewAPI key).
	active := mappings[:0]
	for _, m := range mappings {
		if m.NewAPIKeyID != nil && m.SyncStatus == store.MappingSyncStatusSynced {
			active = append(active, m)
		}
	}
	if len(active) == 0 {
		return nil
	}

	// Collect all platform key IDs for a single batch computation.
	keyIDs := make(map[uuid.UUID]struct{}, len(active))
	for _, m := range active {
		keyIDs[m.PlatformKeyID] = struct{}{}
	}

	// Token is unlimited on NewAPI — no remote quota to sync.
	// Only refresh the local combined_key_remain for gateway precheck.
	updates, err := ComputeGatewaySummaryUpdates(ctx, s.store, keyIDs, s.cfg.Clock())
	if err != nil {
		return err
	}
	if len(updates) == 0 {
		return nil
	}
	summaries, err := s.store.CombinedKeySummaries().UpdateBatch(ctx, updates)
	if err != nil {
		return err
	}
	RefreshCombinedKeySummaries(ctx, s.cache, nil, store.CompanyID(ctx), summaries)
	return nil
}

var _ Rebalancer = (*RebalanceService)(nil)
