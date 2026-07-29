package pricing

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/store"
)

// Service manages model pricing (TJ as SOT) and best-effort syncs to NewAPI.
type Service struct {
	store  store.Store
	client adminport.Port // nullable — NewAPI sync
	cfg    config.Config
}

func NewService(cfg config.Config, st store.Store, client adminport.Port) *Service {
	return &Service{store: st, client: client, cfg: cfg}
}

// catalog.pricing_version is bumped after every price change so CatalogSync
// clients know to re-pull.
const keyPricingVersion = "catalog.pricing_version"

// SetGlobalPrice inserts a global price row and best-effort syncs to NewAPI.
func (s *Service) SetGlobalPrice(ctx context.Context, modelType string, input, output float64, note string) error {
	row := store.ModelPricingRow{
		CompanyID:     s.cfg.TokenJoyCompanyID,
		ModelType:     modelType,
		InputPrice:    input,
		OutputPrice:   output,
		EffectiveFrom: time.Now(),
		Note:          note,
	}
	if err := s.store.ModelPricing().Insert(ctx, row); err != nil {
		return err
	}
	// Best-effort push to NewAPI cache.
	if s.client != nil {
		if err := s.client.UpsertModelRatio(ctx, modelType, input, output); err != nil {
			slog.Warn("newapi pricing sync failed", "model", modelType, "error", err)
		}
	}
	// Bump pricing version for CatalogSync clients.
	_, _ = s.store.SystemSettings().Increment(ctx, keyPricingVersion)
	return nil
}

// SetContractPrice inserts a per-company contract price row.
// Not pushed to NewAPI — gateway has only one global ratio set, no per-company support.
// Contract pricing difference is realized at TJ ingest time (applyTJPricing).
func (s *Service) SetContractPrice(ctx context.Context, companyID uuid.UUID, modelType string, input, output float64, note string) error {
	row := store.ModelPricingRow{
		CompanyID:     companyID,
		ModelType:     modelType,
		InputPrice:    input,
		OutputPrice:   output,
		EffectiveFrom: time.Now(),
		Note:          note,
	}
	if err := s.store.ModelPricing().Insert(ctx, row); err != nil {
		return err
	}
	// Bump pricing version for CatalogSync clients.
	_, _ = s.store.SystemSettings().Increment(ctx, keyPricingVersion)
	return nil
}

// ListGlobalPricing returns current global prices for all models.
func (s *Service) ListGlobalPricing(ctx context.Context) ([]store.ModelPricingRow, error) {
	return s.store.ModelPricing().CurrentPricesBatch(ctx, s.cfg.TokenJoyCompanyID, time.Now())
}

// ListContractPricing returns current contract prices for a company.
func (s *Service) ListContractPricing(ctx context.Context, companyID uuid.UUID) ([]store.ModelPricingRow, error) {
	return s.store.ModelPricing().CurrentPricesBatch(ctx, companyID, time.Now())
}

// PriceHistory returns the full price timeline for a company+model.
func (s *Service) PriceHistory(ctx context.Context, companyID uuid.UUID, modelType string) ([]store.ModelPricingRow, error) {
	return s.store.ModelPricing().History(ctx, companyID, modelType)
}

// FullSyncToNewAPI pushes all current global prices to NewAPI (cron job).
func (s *Service) FullSyncToNewAPI(ctx context.Context) error {
	if s.client == nil {
		return nil
	}
	prices, err := s.store.ModelPricing().CurrentPricesBatch(ctx, s.cfg.TokenJoyCompanyID, time.Now())
	if err != nil {
		return err
	}
	for _, p := range prices {
		if err := s.client.UpsertModelRatio(ctx, p.ModelType, p.InputPrice, p.OutputPrice); err != nil {
			slog.Warn("full sync pricing failed", "model", p.ModelType, "error", err)
		}
	}
	return nil
}

// ResolvePrice returns the effective price for a company+model at a given time.
// Falls back to global price if no contract price exists.
func (s *Service) ResolvePrice(ctx context.Context, companyID uuid.UUID, modelType string, at time.Time) (*store.ModelPricingRow, error) {
	p, err := s.store.ModelPricing().CurrentPrice(ctx, companyID, modelType, at)
	if err != nil {
		return nil, err
	}
	if p != nil {
		return p, nil
	}
	// Fallback to global price.
	return s.store.ModelPricing().CurrentPrice(ctx, s.cfg.TokenJoyCompanyID, modelType, at)
}
