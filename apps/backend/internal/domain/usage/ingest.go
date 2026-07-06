package usage

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/store"
)

type IngestService struct {
	cfg      config.Config
	store    store.Store
	logStore store.LogStore
	notifier notification.Notifier
	logger   *slog.Logger
}

func NewIngestService(
	cfg config.Config,
	st store.Store,
	logStore store.LogStore,
	notifier notification.Notifier,
	logger *slog.Logger,
) *IngestService {
	if logStore == nil {
		logStore = store.NoopLogStore()
	}
	return &IngestService{cfg: cfg, store: st, logStore: logStore, notifier: notifier, logger: logger}
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

	return s.store.WithTx(ctx, func(st store.Store) error {
		inserted, err := st.Ledger().InsertOnConflict(ctx, entry)
		if err != nil || !inserted {
			return err
		}
		if err := Apply(ctx, st, entry); err != nil {
			return err
		}
		return enqueueSideEffects(ctx, st, entry)
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
