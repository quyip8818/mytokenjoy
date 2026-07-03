package usage

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/store"
)

type IngestService struct {
	cfg      config.Config
	store    store.Store
	notifier notification.Notifier
	logger   *slog.Logger
}

func NewIngestService(
	cfg config.Config,
	st store.Store,
	notifier notification.Notifier,
	logger *slog.Logger,
) *IngestService {
	return &IngestService{cfg: cfg, store: st, notifier: notifier, logger: logger}
}

func (s *IngestService) Ingest(ctx context.Context, payload newapi.WebhookLogPayload, source string) error {
	mapping, err := s.store.Relay().FindMappingByNewAPITokenID(ctx, payload.TokenID)
	if err != nil {
		return err
	}
	if mapping == nil {
		s.logger.Warn("ingest rejected: mapping missing", "token_id", payload.TokenID, "log_id", payload.ID)
		return domain.NotFound(fmt.Sprintf("mapping not found for token %d", payload.TokenID))
	}
	ctx, err = s.companyContextFromMapping(ctx, mapping)
	if err != nil {
		return err
	}

	buildInput, err := LoadEntryBuildInput(ctx, s.store, mapping, payload, source)
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

func (s *IngestService) IngestFromOutbox(ctx context.Context, raw json.RawMessage) error {
	var payload newapi.WebhookLogPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}
	return s.Ingest(ctx, payload, types.SourceWebhook)
}

func (s *IngestService) EnqueueFailed(ctx context.Context, payload newapi.WebhookLogPayload, ingestErr error) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return s.store.Relay().EnqueueWebhookOutbox(ctx, store.WebhookOutboxEntry{
		ID:        fmt.Sprintf("wh-%d", time.Now().UnixNano()),
		Payload:   raw,
		Status:    store.OutboxStatusPending,
		NextRetry: time.Now(),
		CreatedAt: time.Now(),
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
