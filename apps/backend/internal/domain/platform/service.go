// Package platform implements the platform administration domain service.
// It encapsulates model catalog management, pricing, company overview aggregation,
// and catalog version control — logic that was previously scattered in the HTTP handler layer.
package platform

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/support/modelcatalog"
)

// Service is the platform administration domain service.
type Service interface {
	// ListModelsWithPricing returns all global platform models enriched with prices from NewAPI.
	ListModelsWithPricing(ctx context.Context) ([]types.ModelInfo, error)
	// CreateModel creates a new platform model and optionally sets its pricing.
	CreateModel(ctx context.Context, input CreateModelInput) (types.ModelInfo, error)
	// UpdateModel updates an existing platform model's metadata.
	UpdateModel(ctx context.Context, id uuid.UUID, input UpdateModelInput) (*types.ModelInfo, error)
	// SetModelPricing sets the pricing for a model by ID.
	SetModelPricing(ctx context.Context, modelID uuid.UUID, pricing PricingInput) error
	// SyncFromNewAPI pulls the model list from NewAPI and upserts into the DB.
	SyncFromNewAPI(ctx context.Context) (int, error)
	// PublishCatalog bumps the models version to trigger downstream syncs.
	PublishCatalog(ctx context.Context) (int, error)
	// ListModelChannels returns NewAPI channels that serve a given model.
	ListModelChannels(ctx context.Context, modelID uuid.UUID) ([]ChannelSummary, error)
	// ListGlobalPricing returns all current global prices from NewAPI.
	ListGlobalPricing(ctx context.Context) ([]PricingEntry, error)
	// SetGlobalPricing sets global pricing for a model type.
	SetGlobalPricing(ctx context.Context, modelType string, pricing PricingInput) error
}

// CreateModelInput is the input for creating a platform model.
type CreateModelInput struct {
	Type            string
	Name            string
	Provider        string
	InputPrice      float64
	OutputPrice     float64
	CacheInputPrice float64
	Capabilities    []string
	MaxContext      int
}

// UpdateModelInput is the input for updating a platform model.
type UpdateModelInput struct {
	Name         *string
	Type         *string
	Provider     *string
	Deprecated   *bool
	Capabilities []string
	MaxContext   *int
}

// PricingInput holds price values for a model.
type PricingInput struct {
	InputPrice      float64
	OutputPrice     float64
	CacheInputPrice float64
}

// PricingEntry is a single model's pricing.
type PricingEntry struct {
	ModelType       string  `json:"modelType"`
	InputPrice      float64 `json:"inputPrice"`
	OutputPrice     float64 `json:"outputPrice"`
	CacheInputPrice float64 `json:"cacheInputPrice"`
}

// ChannelSummary describes a NewAPI channel serving a model.
type ChannelSummary struct {
	Name     string `json:"name"`
	Group    string `json:"group"`
	Priority int    `json:"priority"`
	Weight   int    `json:"weight"`
	Status   int    `json:"status"`
}

// Store is the narrow store surface the platform domain needs.
type Store interface {
	Models() store.ModelsRepository
	SyncVersions() store.SyncVersionRepository
}

type service struct {
	store           Store
	adminPort       adminport.Port
	tokenJoyCompany uuid.UUID
}

// NewService creates a new platform domain service.
func NewService(st Store, port adminport.Port, tokenJoyCompanyID uuid.UUID) Service {
	return &service{
		store:           st,
		adminPort:       port,
		tokenJoyCompany: tokenJoyCompanyID,
	}
}

func (s *service) ListModelsWithPricing(ctx context.Context) ([]types.ModelInfo, error) {
	models, err := s.store.Models().Models(ctx)
	if err != nil {
		return nil, err
	}
	// Filter to global platform models only.
	var global []types.ModelInfo
	for _, m := range models {
		if m.CompanyID == s.tokenJoyCompany {
			global = append(global, m)
		}
	}
	s.mergePricing(ctx, global)
	return global, nil
}

func (s *service) CreateModel(ctx context.Context, input CreateModelInput) (types.ModelInfo, error) {
	if input.Type == "" {
		return types.ModelInfo{}, fmt.Errorf("type required")
	}
	if input.Provider == "" {
		return types.ModelInfo{}, fmt.Errorf("provider required")
	}
	capabilities := input.Capabilities
	if len(capabilities) == 0 {
		capabilities = []string{"chat"}
	}
	maxContext := input.MaxContext
	if maxContext == 0 {
		maxContext = 1000000
	}
	model := types.ModelInfo{
		Provider:     input.Provider,
		Type:         input.Type,
		Name:         input.Name,
		Source:       "platform",
		Deprecated:   false,
		Capabilities: capabilities,
		MaxContext:   maxContext,
	}
	if model.Name == "" {
		model.Name = model.Type
	}
	created, err := s.store.Models().InsertModel(ctx, model)
	if err != nil {
		return types.ModelInfo{}, fmt.Errorf("create model: %w", err)
	}
	// Push price to NewAPI (SOT).
	if input.InputPrice > 0 || input.OutputPrice > 0 {
		_ = s.adminPort.UpsertModelRatio(ctx, input.Type, input.InputPrice, input.OutputPrice, input.CacheInputPrice)
		s.bumpPricing(ctx)
	}
	s.bumpModels(ctx)
	return created, nil
}

func (s *service) UpdateModel(ctx context.Context, id uuid.UUID, input UpdateModelInput) (*types.ModelInfo, error) {
	model, err := s.store.Models().ModelByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if model == nil {
		return nil, fmt.Errorf("model not found")
	}
	if input.Name != nil {
		model.Name = *input.Name
	}
	if input.Type != nil {
		model.Type = *input.Type
	}
	if input.Provider != nil {
		model.Provider = *input.Provider
	}
	if input.Deprecated != nil {
		model.Deprecated = *input.Deprecated
	}
	if input.Capabilities != nil {
		model.Capabilities = input.Capabilities
	}
	if input.MaxContext != nil {
		model.MaxContext = *input.MaxContext
	}
	if err := s.store.Models().UpdateModel(ctx, *model); err != nil {
		return nil, fmt.Errorf("update model: %w", err)
	}
	s.bumpModels(ctx)
	return model, nil
}

func (s *service) SetModelPricing(ctx context.Context, modelID uuid.UUID, pricing PricingInput) error {
	model, err := s.store.Models().ModelByID(ctx, modelID)
	if err != nil || model == nil {
		return fmt.Errorf("model not found")
	}
	if err := s.adminPort.UpsertModelRatio(ctx, model.Type, pricing.InputPrice, pricing.OutputPrice, pricing.CacheInputPrice); err != nil {
		return fmt.Errorf("set pricing: %w", err)
	}
	s.bumpPricing(ctx)
	return nil
}

func (s *service) SyncFromNewAPI(ctx context.Context) (int, error) {
	pricingModels, err := s.adminPort.ListPricingModels(ctx)
	if err != nil {
		return 0, fmt.Errorf("list pricing models: %w", err)
	}
	infos := make([]types.ModelInfo, 0, len(pricingModels))
	for _, pm := range pricingModels {
		if modelcatalog.IsTestCallType(pm.ModelName) {
			continue
		}
		infos = append(infos, pricingModelToModelInfo(pm))
	}
	if err := s.store.Models().SyncFromPlatform(ctx, s.tokenJoyCompany, infos); err != nil {
		return 0, fmt.Errorf("sync models: %w", err)
	}
	s.bumpModels(ctx)
	return len(infos), nil
}

func (s *service) PublishCatalog(ctx context.Context) (int, error) {
	return s.store.SyncVersions().Increment(ctx, store.GlobalSyncVersion, "models")
}

func (s *service) ListModelChannels(ctx context.Context, modelID uuid.UUID) ([]ChannelSummary, error) {
	model, err := s.store.Models().ModelByID(ctx, modelID)
	if err != nil || model == nil {
		return nil, fmt.Errorf("model not found")
	}
	channels, err := s.adminPort.ListChannels(ctx)
	if err != nil {
		return nil, err
	}
	var result []ChannelSummary
	for _, ch := range channels {
		if ch.Models == "" {
			continue
		}
		for _, m := range splitModels(ch.Models) {
			if m == model.Type {
				result = append(result, ChannelSummary{
					Name:     ch.Name,
					Group:    ch.Group,
					Priority: ch.Priority,
					Weight:   ch.Weight,
					Status:   ch.Status,
				})
				break
			}
		}
	}
	return result, nil
}

func (s *service) ListGlobalPricing(ctx context.Context) ([]PricingEntry, error) {
	ratios, err := s.adminPort.ListModelPricing(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]PricingEntry, 0, len(ratios))
	for _, r := range ratios {
		inp, outp, cache := modelcatalog.PriceFromRatio(r.ModelRatio, r.CompletionRatio, r.CacheRatio)
		out = append(out, PricingEntry{ModelType: r.ModelName, InputPrice: inp, OutputPrice: outp, CacheInputPrice: cache})
	}
	return out, nil
}

func (s *service) SetGlobalPricing(ctx context.Context, modelType string, pricing PricingInput) error {
	if err := s.adminPort.UpsertModelRatio(ctx, modelType, pricing.InputPrice, pricing.OutputPrice, pricing.CacheInputPrice); err != nil {
		return err
	}
	s.bumpPricing(ctx)
	return nil
}

// --- internal helpers ---

func (s *service) bumpModels(ctx context.Context) {
	_, _ = s.store.SyncVersions().Increment(ctx, store.GlobalSyncVersion, "models")
}

func (s *service) bumpPricing(ctx context.Context) {
	_, _ = s.store.SyncVersions().Increment(ctx, store.GlobalSyncVersion, "pricing")
}

func (s *service) mergePricing(ctx context.Context, models []types.ModelInfo) {
	ratios, err := s.adminPort.ListModelPricing(ctx)
	if err != nil || len(ratios) == 0 {
		return
	}
	entries := make([]modelcatalog.RatioEntry, len(ratios))
	for i, r := range ratios {
		entries[i] = modelcatalog.RatioEntry{ModelName: r.ModelName, ModelRatio: r.ModelRatio, CompletionRatio: r.CompletionRatio, CacheRatio: r.CacheRatio}
	}
	modelcatalog.MergePricing(models, entries)
}
