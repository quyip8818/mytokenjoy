package budget

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/clock"
	"github.com/tokenjoy/backend/internal/store"
)

// reconcileMaxLedgerEntries is the hard limit on ledger entries loaded per
// reconcile run. If exceeded the job fails and retries (avoids unbounded memory).
const reconcileMaxLedgerEntries = 50000

// ReconcileStore is the narrow store surface the budget reconcile service needs.
type ReconcileStore interface {
	Company() store.CompanyRepository
	BudgetConsumed() store.BudgetConsumedRepository
	Budget() store.BudgetRepository
	Org() store.OrgRepository
	Keys() store.KeysRepository
	Ledger() store.LedgerRepository
	PlatformKeyMappings() store.PlatformKeyMappingRepository
	CombinedKeySummaries() store.CombinedKeySummaryRepository
	WithTx(ctx context.Context, fn func(store.Store) error) error
}

type ReconcileService struct {
	cfg              config.Config
	store            ReconcileStore
	enqueuer         JobEnqueuer
	logger           *slog.Logger
	combinedKeyCache CombinedKeyCache
}

func NewReconcileService(cfg config.Config, st ReconcileStore, enqueuer JobEnqueuer, cache CombinedKeyCache, logger *slog.Logger) *ReconcileService {
	if enqueuer == nil {
		enqueuer = NoopJobEnqueuer
	}
	if cache == nil {
		cache = NoopCombinedKeyCache
	}
	return &ReconcileService{cfg: cfg, store: st, enqueuer: enqueuer, logger: logger, combinedKeyCache: cache}
}

func (s *ReconcileService) RunCompany(ctx context.Context, companyID uuid.UUID) error {
	co, err := s.store.Company().GetByID(ctx, companyID)
	if err != nil {
		return err
	}
	if co == nil {
		return nil
	}
	ctx = company.WithContext(ctx, company.ContextFromStore(*co))

	var summaries []store.CombinedKeySummary
	repaired := false

	err = s.store.WithTx(ctx, func(tx store.Store) error {
		// 1. Acquire advisory lock (coordinates with budget management writes).
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}

		// 2. Lock company row — serializes with Ingest.
		if _, err := tx.Company().LockForUpdate(ctx, companyID); err != nil {
			return err
		}

		// 3. Read window ledger INSIDE the lock.
		since := ReconcileWindowStart(clock.NowUTC(s.cfg.Clock()))
		entries, err := tx.Ledger().ListCallSettledSince(ctx, since)
		if err != nil {
			return err
		}
		if len(entries) > reconcileMaxLedgerEntries {
			return fmt.Errorf("budget reconcile: company %d has %d entries (limit %d), skipping",
				companyID, len(entries), reconcileMaxLedgerEntries)
		}

		// 4. Compute expected consumed per axis using each entry's OccurredAt
		//    for period attribution (matches what Ingest wrote at that time).
		nodes := tx.Org().Nodes()
		expected, err := expectedConsumedByEntryTime(ctx, nodes, entries)
		if err != nil {
			return err
		}

		// 5. Load actual consumed for the period keys present in expected.
		periodKeys := CollectPeriodKeys(expected)
		consumedRepo := tx.BudgetConsumed()

		// Build actual map from all axis kinds.
		actual := make(map[AxisKey]int64)
		for _, axisKind := range []string{store.AxisKindPlatformKey, store.AxisKindMember, store.AxisKindProject} {
			byPeriod, err := consumedRepo.ListConsumedByPeriods(ctx, axisKind, periodKeys)
			if err != nil {
				return err
			}
			for pk, axisMap := range byPeriod {
				for axisID, consumed := range axisMap {
					actual[AxisKey{Kind: axisKind, AxisID: axisID, PeriodKey: pk}] = consumed
				}
			}
		}

		// 6. Diff expected vs actual.
		repairedAxes := make(map[AxisKey]struct{})

		// Fix drift and missing rows.
		for key, want := range expected {
			got := actual[key]
			if !ConsumedDrift(want, got) {
				continue
			}
			if err := consumedRepo.SetConsumed(ctx, key.Kind, key.AxisID, key.PeriodKey, want); err != nil {
				return err
			}
			repaired = true
			repairedAxes[key] = struct{}{}
			delete(actual, key) // processed
			if s.logger != nil {
				s.logger.Warn("budget reconcile drift repaired",
					"company_id", companyID,
					"axis_kind", key.Kind,
					"axis_id", key.AxisID,
					"period_key", key.PeriodKey,
					"expected", want,
					"actual", got,
				)
			}
		}

		// Clear excess rows (in actual but not in expected → stale).
		for key, got := range actual {
			if _, inExpected := expected[key]; inExpected {
				continue
			}
			if got == 0 {
				continue // already zero, no-op
			}
			if err := consumedRepo.SetConsumed(ctx, key.Kind, key.AxisID, key.PeriodKey, 0); err != nil {
				return err
			}
			repaired = true
			repairedAxes[key] = struct{}{}
			if s.logger != nil {
				s.logger.Warn("budget reconcile excess row cleared",
					"company_id", companyID,
					"axis_kind", key.Kind,
					"axis_id", key.AxisID,
					"period_key", key.PeriodKey,
					"was", got,
				)
			}
		}

		if !repaired {
			return nil
		}

		// 7. Recompute combined key remain for affected platform keys.
		affectedKeys, err := AffectedPlatformKeyIDs(ctx, tx, repairedAxes)
		if err != nil {
			return err
		}
		if len(affectedKeys) > 0 {
			// Lock platform_keys in stable order before absolute recompute.
			sortedKeyIDs := SortedKeys(affectedKeys)
			if err := tx.CombinedKeySummaries().LockPlatformKeysForUpdate(ctx, sortedKeyIDs); err != nil {
				return err
			}
			updates, err := ComputeGatewaySummaryUpdates(ctx, tx, affectedKeys, s.cfg.Clock())
			if err != nil {
				return err
			}
			if len(updates) > 0 {
				summaries, err = tx.CombinedKeySummaries().UpdateBatch(ctx, updates)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if !repaired {
		return nil
	}

	// Post-commit: refresh cache and trigger rebalance.
	RefreshCombinedKeySummaries(ctx, s.combinedKeyCache, s.logger, companyID, summaries)
	return s.enqueuer.InsertRebalance(ctx, companyID, store.RebalanceAxisCompany, companyID)
}

// expectedConsumedByEntryTime computes expected consumed aggregation using each
// entry's OccurredAt to determine its budget period — matching what Ingest wrote.
func expectedConsumedByEntryTime(ctx context.Context, nodes store.OrgNodeRepository, entries []types.UsageLedgerEntry) (map[AxisKey]int64, error) {
	acc := make(map[AxisKey]int64)
	for _, entry := range entries {
		// Use OccurredAt as the reference time for open period — this matches
		// what Ingest computed at write time (Clock ≈ OccurredAt for fresh entries).
		open, err := pkgbudget.OpenDepartmentPeriodAt(ctx, nodes, entry.DepartmentID, entry.OccurredAt)
		if err != nil {
			return nil, err
		}
		deltas, err := ConsumptionDeltas(ctx, nodes, entry, open)
		if err != nil {
			return nil, err
		}
		for _, d := range deltas {
			key := AxisKey{Kind: d.Kind, AxisID: d.AxisID, PeriodKey: d.PeriodKey}
			acc[key] += d.Amount
		}
	}
	return acc, nil
}

func CollectPeriodKeys(expected map[AxisKey]int64) []string {
	set := make(map[string]struct{})
	for key := range expected {
		set[key.PeriodKey] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for pk := range set {
		out = append(out, pk)
	}
	return out
}

func SortedKeys(m map[uuid.UUID]struct{}) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].String() < out[j].String()
	})
	return out
}

func ReconcileWindowStart(now time.Time) time.Time {
	return now.AddDate(0, -2, 0)
}
