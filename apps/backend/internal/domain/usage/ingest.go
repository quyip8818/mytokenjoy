package usage

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	billinglot "github.com/tokenjoy/backend/internal/domain/billing/lot"
	"github.com/tokenjoy/backend/internal/domain/company"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

type IngestService struct {
	cfg      config.Config
	store    store.Store
	logStore store.LogStore
	logger   *slog.Logger
	enqueuer IngestJobEnqueuer
}

func NewIngestService(
	cfg config.Config,
	st store.Store,
	logStore store.LogStore,
	logger *slog.Logger,
	enqueuer IngestJobEnqueuer,
) *IngestService {
	if logStore == nil {
		logStore = store.NoopLogStore()
	}
	if enqueuer == nil {
		enqueuer = noopIngestEnqueuer{}
	}
	return &IngestService{
		cfg: cfg, store: st, logStore: logStore, logger: logger,
		enqueuer: enqueuer,
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

func (s *IngestService) IngestRaw(ctx context.Context, raw store.RawConsumeLog, source string) error {
	mapping, err := s.store.PlatformKeyMappings().FindMappingByNewAPIKeyID(ctx, raw.TokenID)
	if err != nil {
		return err
	}
	if mapping == nil {
		s.logger.Warn("ingest rejected: mapping missing", "token_id", raw.TokenID, "log_id", raw.ID)
		return domain.NotFound(fmt.Sprintf("mapping not found for token %d", raw.TokenID))
	}
	ctx, err = s.companyContextFromMapping(ctx, mapping)
	if err != nil {
		return err
	}

	buildInput, err := LoadEntryBuildInput(ctx, s.store, mapping, raw, source)
	if err != nil {
		return err
	}
	entry, err := BuildCallSettledEntry(buildInput)
	if err != nil {
		return err
	}
	nodes := s.store.Org().Nodes()
	occurrence, err := pkgbudget.OccurrenceDepartmentPeriod(ctx, nodes, entry.DepartmentID, entry.OccurredAt)
	if err != nil {
		return err
	}
	entry.PeriodKey = occurrence.String()

	return s.store.WithTx(ctx, func(st store.Store) error {
		if err := st.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}
		if exists, err := st.Ledger().ExistsIdempotency(ctx, entry.IdempotencyKey); err != nil {
			return err
		} else if exists {
			return nil
		}
		segs, err := billinglot.ConsumeLots(ctx, st, company.CompanyID(ctx), entry.Amount)
		if err != nil {
			return err
		}
		ledgerEntries := billinglot.LedgerSegmentsFromEntry(entry, segs)
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
		if err := s.enqueuer.EnqueueAfterIngest(ctx, tx, company.CompanyID(ctx)); err != nil {
			return err
		}
		return nil
	})
}

func (s *IngestService) companyContextFromMapping(ctx context.Context, mapping *store.PlatformKeyMapping) (context.Context, error) {
	companyCtx := company.Context{CompanyID: mapping.CompanyID}
	co, err := s.store.Company().GetByID(ctx, mapping.CompanyID)
	if err != nil {
		return nil, err
	}
	if co != nil {
		companyCtx.Slug = co.Slug
		companyCtx.Status = co.Status
		if co.NewAPIWalletUserID != nil {
			companyCtx.NewAPIWalletUserID = *co.NewAPIWalletUserID
		}
	}
	return company.WithContext(ctx, companyCtx), nil
}

var _ Ingestor = (*IngestService)(nil)
