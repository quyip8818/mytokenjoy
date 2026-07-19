package usage

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

type IngestService struct {
	cfg         config.Config
	store       store.Store
	logStore    store.LogStore
	logger      *slog.Logger
	enqueuer    IngestJobEnqueuer
	notifier    types.Notifier
	budgetOps   BudgetOps
	lotConsumer LotConsumer
}

func NewIngestService(
	cfg config.Config,
	st store.Store,
	logStore store.LogStore,
	logger *slog.Logger,
	enqueuer IngestJobEnqueuer,
	notifier types.Notifier,
	budgetOps BudgetOps,
	lotConsumer LotConsumer,
) *IngestService {
	if logStore == nil {
		logStore = store.NoopLogStore()
	}
	if enqueuer == nil {
		enqueuer = noopIngestEnqueuer{}
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &IngestService{
		cfg: cfg, store: st, logStore: logStore, logger: logger,
		enqueuer: enqueuer, notifier: notifier,
		budgetOps: budgetOps, lotConsumer: lotConsumer,
	}
}

type noopIngestEnqueuer struct{}

func (noopIngestEnqueuer) EnqueueAfterIngest(context.Context, store.Tx, uuid.UUID, *IngestEffects) error {
	return nil
}

// IngestEffects captures the side-effects produced during the ingest transaction
// so post-commit logic (alerts, cache refresh) can act on them without re-querying.
type IngestEffects struct {
	TouchedDepartments map[uuid.UUID]struct{}
	Summaries          []store.CombinedKeySummary
	PlatformKeyID      uuid.UUID
	OverrunPayload     json.RawMessage // pre-built overrun job payload
}

func (s *IngestService) IngestByLogID(ctx context.Context, logID int64, source string) error {
	raw, err := s.logStore.GetConsumeLogByID(ctx, logID)
	if err != nil {
		return err
	}
	return s.IngestRaw(ctx, *raw, source)
}

// CompanyIDsByLogID resolves company IDs for a batch of consume log IDs (best-effort).
// Missing logs or mappings are omitted from the result map.
func (s *IngestService) CompanyIDsByLogID(ctx context.Context, logIDs []int64) (map[int64]uuid.UUID, error) {
	if len(logIDs) == 0 {
		return nil, nil
	}
	logs, err := s.logStore.GetConsumeLogsByIDs(ctx, logIDs)
	if err != nil {
		return nil, err
	}
	tokenIDs := make([]int64, 0, len(logs))
	seenToken := make(map[int64]struct{}, len(logs))
	for _, raw := range logs {
		if _, ok := seenToken[raw.TokenID]; ok {
			continue
		}
		seenToken[raw.TokenID] = struct{}{}
		tokenIDs = append(tokenIDs, raw.TokenID)
	}
	mappings, err := s.store.PlatformKeyMappings().ListMappingsByNewAPIKeyIDs(ctx, tokenIDs)
	if err != nil {
		return nil, err
	}
	companyByToken := make(map[int64]uuid.UUID, len(mappings))
	for _, m := range mappings {
		if m.NewAPIKeyID == nil {
			continue
		}
		companyByToken[*m.NewAPIKeyID] = m.CompanyID
	}
	out := make(map[int64]uuid.UUID, len(logs))
	for _, raw := range logs {
		if companyID, ok := companyByToken[raw.TokenID]; ok {
			out[raw.ID] = companyID
		}
	}
	return out, nil
}

func (s *IngestService) IngestRaw(ctx context.Context, raw store.RawConsumeLog, source string) error {
	mapping, err := s.store.PlatformKeyMappings().FindMappingByNewAPIKeyID(ctx, raw.TokenID)
	if err != nil {
		return err
	}
	if mapping == nil {
		return domain.NotFound(fmt.Sprintf("mapping not found for token %d", raw.TokenID))
	}
	ctx, err = s.companyContextFromMapping(ctx, mapping)
	if err != nil {
		return err
	}

	snap, err := LoadEntryBuildSnapshot(ctx, s.store)
	if err != nil {
		return err
	}
	buildInput, err := LoadEntryBuildInput(ctx, s.store, mapping, raw, source, snap)
	if err != nil {
		return err
	}
	entry, err := BuildCallSettledEntry(buildInput)
	if err != nil {
		return err
	}
	occurrence, err := pkgbudget.OccurrenceDepartmentPeriodFromTree(snap.OrgTree, entry.DepartmentID, entry.OccurredAt)
	if err != nil {
		return err
	}
	entry.PeriodKey = occurrence.String()

	companyID := company.CompanyID(ctx)
	var consumeResult LotConsumeResult
	var effects IngestEffects
	err = s.store.WithTx(ctx, func(st store.Store) error {
		// 1. Lock company row — serializes all ingest + reconcile for this company.
		co, lockErr := st.Company().LockForUpdate(ctx, companyID)
		if lockErr != nil {
			return lockErr
		}
		if co == nil {
			return fmt.Errorf("ingest: company %s not found", companyID)
		}

		// 2. Idempotency check AFTER lock — guarantees zero side-effects on duplicate.
		if exists, err := st.Ledger().ExistsIdempotency(ctx, entry.IdempotencyKey); err != nil {
			return err
		} else if exists {
			return nil
		}

		// 3. Consume lots (company already locked).
		result, err := s.lotConsumer.ConsumeLotsLocked(ctx, st, co, entry.Amount)
		if err != nil {
			return err
		}
		consumeResult = result

		// 4. Insert ledger segments.
		ledgerEntries := s.lotConsumer.LedgerSegmentsFromEntry(entry, result.Segments)
		inserted, err := st.Ledger().InsertSegments(ctx, ledgerEntries)
		if err != nil {
			return err
		}
		if inserted == 0 {
			return fmt.Errorf("ingest: ledger insert returned zero rows")
		}

		// 5. Write budget_consumed — batch UPSERT for open-budget period axes.
		open, err := pkgbudget.OpenDepartmentPeriod(ctx, st.Org().Nodes(), entry.DepartmentID, s.cfg.Clock())
		if err != nil {
			return err
		}
		deltas, err := s.budgetOps.ConsumptionDeltas(ctx, entry, open)
		if err != nil {
			return err
		}
		storeDeltas := make([]store.ConsumedDelta, len(deltas))
		for i, d := range deltas {
			storeDeltas[i] = store.ConsumedDelta{
				AxisKind:  d.Kind,
				AxisID:    d.AxisID,
				PeriodKey: d.PeriodKey,
				Amount:    d.Amount,
			}
		}
		if err := st.BudgetConsumed().IncrementConsumedBatch(ctx, storeDeltas); err != nil {
			return err
		}

		// 6. Decrement combined_key_remain.
		var summaries []store.CombinedKeySummary
		if entry.PlatformKeyID != uuid.Nil {
			decrements := map[uuid.UUID]int64{entry.PlatformKeyID: entry.Amount}
			summaries, err = st.CombinedKeySummaries().DecrementBatch(ctx, decrements)
			if err != nil {
				return err
			}
			// Handle NULL / missing key — absolute recompute if key was not decremented.
			if len(summaries) == 0 {
				summaries, err = s.absoluteRecompute(ctx, st, entry.PlatformKeyID)
				if err != nil {
					// Don't fail the ingest — log and treat as Unknown for overrun gate.
					// Use empty (non-nil) slice to signal Unknown state to ShouldEnqueueOverrun.
					s.logger.Warn("combined key absolute recompute failed",
						"platform_key_id", entry.PlatformKeyID, "error", err)
					summaries = []store.CombinedKeySummary{}
				}
			}
		}
		effects.Summaries = summaries
		effects.TouchedDepartments = map[uuid.UUID]struct{}{entry.DepartmentID: {}}
		effects.PlatformKeyID = entry.PlatformKeyID
		effects.OverrunPayload = OverrunPayloadFromEntry(entry, open.String())

		// 7. Enqueue side-effect jobs (dashboard, wallet, conditional overrun).
		tx, ok := st.(store.Tx)
		if !ok {
			return fmt.Errorf("ingest: transaction store required")
		}
		return s.enqueuer.EnqueueAfterIngest(ctx, tx, companyID, &effects)
	})
	if err != nil {
		return err
	}

	// --- Post-commit (best-effort) ---

	// Refresh Redis combined key cache.
	s.budgetOps.RefreshCombinedKeySummaries(ctx, companyID, effects.Summaries)

	// Check budget alert thresholds for touched departments.
	s.budgetOps.CheckBudgetAlerts(ctx, s.store, companyID, effects.TouchedDepartments)

	// Notify overdraft expansion.
	if consumeResult.OverdraftUsed && s.notifier != nil {
		_ = s.notifier.Send(ctx, types.Notification{
			EventType: types.NotificationEventOverdraftExpanded,
			Payload: map[string]any{
				"companyId":      companyID,
				"overdraftDelta": consumeResult.OverdraftDelta,
			},
		})
	}
	return nil
}

// absoluteRecompute handles the rare case where DecrementBatch did not update a key
// (combined_key_remain was NULL). It locks the platform_keys row, recomputes the
// combined remain from budget context, and writes the absolute value.
func (s *IngestService) absoluteRecompute(ctx context.Context, st store.Store, platformKeyID uuid.UUID) ([]store.CombinedKeySummary, error) {
	keyIDs := []uuid.UUID{platformKeyID}
	if err := st.CombinedKeySummaries().LockPlatformKeysForUpdate(ctx, keyIDs); err != nil {
		return nil, err
	}
	keySet := make(map[uuid.UUID]struct{}, 1)
	keySet[platformKeyID] = struct{}{}
	updates, err := s.budgetOps.ComputeGatewaySummaryUpdates(ctx, st, keySet, s.cfg.Clock())
	if err != nil {
		return nil, err
	}
	if len(updates) == 0 {
		// Unconstrained — no budget limits, leave NULL.
		return nil, nil
	}
	return st.CombinedKeySummaries().UpdateBatch(ctx, updates)
}

// OverrunPayloadFromEntry creates the overrun job payload matching the existing OverrunService schema.
func OverrunPayloadFromEntry(entry types.UsageLedgerEntry, periodKey string) json.RawMessage {
	payload := map[string]any{
		"departmentId":  entry.DepartmentID,
		"platformKeyId": entry.PlatformKeyID,
		"periodKey":     periodKey,
	}
	if entry.MemberID != nil {
		payload["memberId"] = *entry.MemberID
	}
	if entry.ProjectID != nil {
		payload["projectId"] = *entry.ProjectID
	}
	raw, _ := json.Marshal(payload)
	return raw
}

// ShouldEnqueueOverrun decides whether an overrun job is needed based on combined remain.
func ShouldEnqueueOverrun(summaries []store.CombinedKeySummary, platformKeyID uuid.UUID) bool {
	if platformKeyID == uuid.Nil {
		return false
	}
	// If absolute recompute returned nil (Unconstrained), skip overrun.
	if summaries == nil {
		return false
	}
	for _, s := range summaries {
		if s.PlatformKeyID == platformKeyID {
			return s.Remain <= 0
		}
	}
	// Key not in summaries = Unknown state → enqueue overrun for safety.
	return true
}

// OverrunPayloadFromEffects returns the pre-built overrun payload, or nil if not applicable.
func OverrunPayloadFromEffects(effects *IngestEffects) json.RawMessage {
	if effects == nil || len(effects.OverrunPayload) == 0 {
		return nil
	}
	return effects.OverrunPayload
}

func (s *IngestService) companyContextFromMapping(ctx context.Context, mapping *store.PlatformKeyMapping) (context.Context, error) {
	co, err := s.store.Company().GetByID(ctx, mapping.CompanyID)
	if err != nil {
		return nil, err
	}
	if co == nil {
		return company.WithContext(ctx, company.Context{CompanyID: mapping.CompanyID}), nil
	}
	return company.WithContext(ctx, company.ContextFromStore(*co)), nil
}

var _ Ingestor = (*IngestService)(nil)
