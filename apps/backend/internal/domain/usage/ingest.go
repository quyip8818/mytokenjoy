package usage

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	billinglot "github.com/tokenjoy/backend/internal/domain/billing/lot"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

type IngestService struct {
	cfg      config.Config
	store    store.Store
	logStore store.LogStore
	logger   *slog.Logger
	enqueuer IngestJobEnqueuer
	notifier types.Notifier
}

func NewIngestService(
	cfg config.Config,
	st store.Store,
	logStore store.LogStore,
	logger *slog.Logger,
	enqueuer IngestJobEnqueuer,
	notifier types.Notifier,
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
	}
}

type noopIngestEnqueuer struct{}

func (noopIngestEnqueuer) EnqueueAfterIngest(context.Context, store.Tx, int64) error { return nil }

func (s *IngestService) IngestByLogID(ctx context.Context, logID int64, source string) error {
	raw, err := s.logStore.GetConsumeLogByID(ctx, logID)
	if err != nil {
		return err
	}
	return s.IngestRaw(ctx, *raw, source)
}

// CompanyIDsByLogID resolves company IDs for a batch of consume log IDs (best-effort).
// Missing logs or mappings are omitted from the result map.
func (s *IngestService) CompanyIDsByLogID(ctx context.Context, logIDs []int64) (map[int64]int64, error) {
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
	companyByToken := make(map[int64]int64, len(mappings))
	for _, m := range mappings {
		if m.NewAPIKeyID == nil {
			continue
		}
		companyByToken[*m.NewAPIKeyID] = m.CompanyID
	}
	out := make(map[int64]int64, len(logs))
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
	var consumeResult billinglot.ConsumeResult
	err = s.store.WithTx(ctx, func(st store.Store) error {
		if exists, err := st.Ledger().ExistsIdempotency(ctx, entry.IdempotencyKey); err != nil {
			return err
		} else if exists {
			return nil
		}
		result, err := billinglot.ConsumeLots(ctx, st, companyID, entry.Amount)
		if err != nil {
			return err
		}
		consumeResult = result
		ledgerEntries := billinglot.LedgerSegmentsFromEntry(entry, result.Segments)
		inserted, err := st.Ledger().InsertSegments(ctx, ledgerEntries)
		if err != nil {
			return err
		}
		if inserted == 0 {
			return fmt.Errorf("ingest: ledger insert returned zero rows")
		}
		tx, ok := st.(store.Tx)
		if !ok {
			return fmt.Errorf("ingest: transaction store required")
		}
		return s.enqueuer.EnqueueAfterIngest(ctx, tx, companyID)
	})
	if err != nil {
		return err
	}

	// Notify outside the transaction to avoid blocking commit on notification failures.
	if consumeResult.OverdraftUsed && s.notifier != nil {
		_ = s.notifier.Send(ctx, types.Notification{
			EventType: types.NotificationEventOverdraftExpanded,
			Recipient: fmt.Sprintf("company:%d", companyID),
			Payload: map[string]any{
				"companyId":      companyID,
				"overdraftDelta": consumeResult.OverdraftDelta,
			},
		})
	}
	return nil
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
