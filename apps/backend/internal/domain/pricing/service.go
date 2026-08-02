package pricing

import (
	"context"
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

// Service manages global pricing (models.input_price/output_price) and best-effort syncs to NewAPI.
// Contract/discount pricing is handled by model_discount, not this service.
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

// SetGlobalPrice updates models.input_price/output_price and best-effort syncs to NewAPI.
func (s *Service) SetGlobalPrice(ctx context.Context, modelType string, input, output float64) error {
	if err := s.store.Models().UpdatePrice(ctx, s.cfg.TokenJoyCompanyID, modelType, input, output); err != nil {
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

// ListGlobalPricing returns current global prices for all models (from models table).
func (s *Service) ListGlobalPricing(ctx context.Context) ([]types.ModelInfo, error) {
	all, err := s.store.Models().ModelsByCompany(ctx, s.cfg.TokenJoyCompanyID)
	if err != nil {
		return nil, err
	}
	// Only return models with pricing set.
	var out []types.ModelInfo
	for _, m := range all {
		if m.InputPrice > 0 || m.OutputPrice > 0 {
			out = append(out, m)
		}
	}
	return out, nil
}

// FullSyncToNewAPI pushes all current global prices to NewAPI (periodic job).
func (s *Service) FullSyncToNewAPI(ctx context.Context) error {
	if s.client == nil {
		return nil
	}
	models, err := s.store.Models().ModelsByCompany(ctx, s.cfg.TokenJoyCompanyID)
	if err != nil {
		return err
	}
	for _, m := range models {
		if m.InputPrice <= 0 && m.OutputPrice <= 0 {
			continue
		}
		if err := s.client.UpsertModelRatio(ctx, m.Type, m.InputPrice, m.OutputPrice); err != nil {
			slog.Warn("full sync pricing failed", "model", m.Type, "error", err)
		}
	}
	return nil
}
