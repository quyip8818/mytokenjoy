package usage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/infra/notification"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

type IngestService struct {
	cfg               config.Config
	store             store.Store
	logStore          store.LogStore
	notifier          notification.Notifier
	logger            *slog.Logger
	enqueueWalletSync func(ctx context.Context, companyID int64) error
}

func NewIngestService(
	cfg config.Config,
	st store.Store,
	logStore store.LogStore,
	notifier notification.Notifier,
	logger *slog.Logger,
	enqueueWalletSync func(ctx context.Context, companyID int64) error,
) *IngestService {
	if logStore == nil {
		logStore = store.NoopLogStore()
	}
	return &IngestService{
		cfg: cfg, store: st, logStore: logStore, notifier: notifier, logger: logger,
		enqueueWalletSync: enqueueWalletSync,
	}
}

func (s *IngestService) IngestByLogID(ctx context.Context, logID int64, source string) error {
	raw, err := s.logStore.GetConsumeLogByID(ctx, logID)
	if err != nil {
		return err
	}
	return s.IngestRaw(ctx, *raw, source)
}

func (s *IngestService) IngestRaw(ctx context.Context, raw store.RawConsumeLog, source string) error {
	mapping, err := s.store.Relay().FindMappingByNewAPITokenID(ctx, raw.TokenID)
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
	ledgerPeriodKey, err := pkgbudget.DepartmentPeriodKey(ctx, nodes, entry.DepartmentID, entry.OccurredAt)
	if err != nil {
		return err
	}
	snapshotPeriodKey, err := pkgbudget.DepartmentPeriodKey(ctx, nodes, entry.DepartmentID, time.Now().UTC())
	if err != nil {
		return err
	}
	entry.PeriodKey = ledgerPeriodKey

	return s.store.WithTx(ctx, func(st store.Store) error {
		if exists, err := st.Ledger().ExistsIdempotency(ctx, entry.IdempotencyKey); err != nil {
			return err
		} else if exists {
			return nil
		}
		segs, err := AllocateConsumptionLots(ctx, st, company.CompanyID(ctx), entry.Amount)
		if err != nil {
			return err
		}
		ledgerEntries := LedgerSegmentsFromEntry(entry, segs)
		inserted, err := st.Ledger().InsertSegments(ctx, ledgerEntries)
		if err != nil || inserted == 0 {
			return err
		}
		if err := Apply(ctx, st, entry, snapshotPeriodKey); err != nil {
			return err
		}
		if err := enqueueSideEffects(ctx, st, entry); err != nil {
			return err
		}
		if s.enqueueWalletSync != nil {
			_ = s.enqueueWalletSync(ctx, company.CompanyID(ctx))
		}
		return nil
	})
}

func (s *IngestService) companyContextFromMapping(ctx context.Context, mapping *store.RelayMapping) (context.Context, error) {
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
