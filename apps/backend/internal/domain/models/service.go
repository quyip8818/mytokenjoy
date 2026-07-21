package models

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
	"github.com/tokenjoy/backend/internal/pkg/newapiunits"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	ListModels(ctx context.Context) ([]types.ModelInfo, error)
	ListModelsWithPricing(ctx context.Context) ([]types.ModelInfoWithPricing, error)
	CreateModel(ctx context.Context, input types.CreateModelInput) (types.ModelInfo, error)
	UpdateModel(ctx context.Context, id uuid.UUID, input types.UpdateModelInput) (types.ModelInfo, error)
	DeleteModel(ctx context.Context, id uuid.UUID) error
	ToggleModel(ctx context.Context, id uuid.UUID, enabled bool) error
	ListRoutingRules(ctx context.Context) ([]types.RoutingRule, error)
	ResolveRouting(ctx context.Context, deptID uuid.UUID) (types.ResolvedWhitelist, error)
	UpdateRoutingRule(ctx context.Context, id uuid.UUID, input types.UpdateRoutingRuleInput) (types.RoutingRule, error)
}

// Store is the narrow store surface the models domain needs.
type Store interface {
	Models() store.ModelsRepository
	Org() store.OrgRepository
}

type service struct {
	cfg              config.Config
	store            Store
	delayer          common.Delayer
	client           adminport.Port
	cacheInvalidator types.PrecheckCacheInvalidator
}

func NewService(cfg config.Config, st Store, client adminport.Port, cacheInvalidator types.PrecheckCacheInvalidator, delayer common.Delayer) Service {
	if cacheInvalidator == nil {
		cacheInvalidator = types.NoopPrecheckCacheInvalidator{}
	}
	return &service{
		cfg:              cfg,
		store:            st,
		delayer:          delayer,
		client:           client,
		cacheInvalidator: cacheInvalidator,
	}
}

func (s *service) ListModels(ctx context.Context) ([]types.ModelInfo, error) {
	return s.store.Models().Models(ctx)
}

func (s *service) CreateModel(ctx context.Context, input types.CreateModelInput) (types.ModelInfo, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.ModelInfo{}, err
	}
	name := input.Name
	if name == "" {
		name = input.Type
	}
	maxContext := input.MaxContext
	if maxContext == 0 {
		maxContext = 1000000
	}
	capabilities := input.Capabilities
	if len(capabilities) == 0 {
		capabilities = []string{"chat"}
	}
	var apiKey *string
	if input.ApiKey != "" {
		apiKey = &input.ApiKey
	}
	var endpointModelName *string
	if input.EndpointModelName != "" {
		endpointModelName = &input.EndpointModelName
	}
	model := types.ModelInfo{
		Provider:          types.ProviderCustom,
		Type:              input.Type,
		Name:              name,
		Description:       "",
		Endpoint:          &input.BaseURL,
		ApiKey:            apiKey,
		EndpointModelName: endpointModelName,
		MaxContext:        maxContext,
		MaxTokens:         input.MaxTokens,
		Enabled:           true,
		Capabilities:      capabilities,
	}
	if err := s.validateModelProviderTypeAvailable(ctx, types.ProviderCustom, input.Type); err != nil {
		return types.ModelInfo{}, err
	}
	created, err := s.store.Models().InsertModel(ctx, model)
	if err != nil {
		return types.ModelInfo{}, mapModelPersistError(err)
	}
	// Write pricing to NewAPI (SOT for all model pricing).
	// Fail the creation if pricing sync fails — a model without pricing
	// would cause consume_log quota to use the default ratio (likely 0).
	if s.client != nil && (input.InputPrice > 0 || input.OutputPrice > 0) {
		if err := s.client.UpsertModelRatio(ctx, input.Type, input.InputPrice, input.OutputPrice); err != nil {
			// Best-effort rollback: remove the local model so we don't leave
			// an orphan without pricing.
			_ = s.store.Models().DeleteModel(ctx, created.ID)
			return types.ModelInfo{}, fmt.Errorf("create model pricing: %w", err)
		}
	}
	return created, nil
}

func (s *service) UpdateModel(ctx context.Context, id uuid.UUID, input types.UpdateModelInput) (types.ModelInfo, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.ModelInfo{}, err
	}
	existing, err := s.requireTenantModel(ctx, id)
	if err != nil {
		return types.ModelInfo{}, err
	}
	if input.Type != nil && *input.Type != existing.Type {
		if err := s.validateModelProviderTypeAvailable(ctx, existing.Provider, *input.Type); err != nil {
			return types.ModelInfo{}, err
		}
		existing.Type = *input.Type
	}
	if input.Name != nil {
		existing.Name = *input.Name
	}
	if input.Description != nil {
		existing.Description = *input.Description
	}
	if input.Endpoint != nil && existing.IsCustom() {
		existing.Endpoint = input.Endpoint
	}
	if input.ApiKey != nil && existing.IsCustom() {
		existing.ApiKey = input.ApiKey
	}
	if input.EndpointModelName != nil && existing.IsCustom() {
		existing.EndpointModelName = input.EndpointModelName
	}
	// Write pricing to NewAPI if changed. Resolve current prices for merge.
	if (input.InputPrice != nil || input.OutputPrice != nil) && s.client != nil {
		var curInput, curOutput float64
		if pricing, err := s.client.ListModelPricing(ctx); err == nil {
			for _, p := range pricing {
				if p.ModelName == existing.Type {
					curInput, curOutput = newapiunits.PriceFromRatio(p.ModelRatio, p.CompletionRatio)
					break
				}
			}
		}
		if input.InputPrice != nil {
			curInput = *input.InputPrice
		}
		if input.OutputPrice != nil {
			curOutput = *input.OutputPrice
		}
		if err := s.client.UpsertModelRatio(ctx, existing.Type, curInput, curOutput); err != nil {
			return types.ModelInfo{}, fmt.Errorf("update model pricing: %w", err)
		}
	}
	if input.MaxContext != nil {
		existing.MaxContext = *input.MaxContext
	}
	if input.MaxTokens != nil {
		existing.MaxTokens = *input.MaxTokens
	}
	if input.Capabilities != nil {
		existing.Capabilities = append([]string{}, input.Capabilities...)
	}
	if err := s.store.Models().UpdateModel(ctx, *existing); err != nil {
		return types.ModelInfo{}, mapModelPersistError(err)
	}
	return *existing, nil
}

func (s *service) DeleteModel(ctx context.Context, id uuid.UUID) error {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return err
	}
	if _, err := s.requireTenantModel(ctx, id); err != nil {
		return err
	}
	return s.store.Models().DeleteModel(ctx, id)
}

func (s *service) ToggleModel(ctx context.Context, id uuid.UUID, enabled bool) error {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return err
	}
	model, err := s.store.Models().ModelByID(ctx, id)
	if err != nil {
		return err
	}
	if model == nil {
		return domain.NotFound("Not found")
	}

	// Global (builtin) model: create a tenant-level override copy with the desired enabled state.
	// DedupeEffective will pick the tenant copy over the global one.
	if model.CompanyID == s.cfg.TokenJoyCompanyID {
		override := *model
		override.Enabled = enabled
		if _, err := s.store.Models().InsertModel(ctx, override); err != nil {
			// If tenant override already exists (duplicate provider+type), update it instead.
			existing, findErr := s.store.Models().ModelByProviderType(ctx, model.Provider, model.Type)
			if findErr != nil || existing == nil {
				return mapModelPersistError(err)
			}
			existing.Enabled = enabled
			if updateErr := s.store.Models().UpdateModel(ctx, *existing); updateErr != nil {
				return mapModelPersistError(updateErr)
			}
		}
		return nil
	}

	model.Enabled = enabled
	if err := s.store.Models().UpdateModel(ctx, *model); err != nil {
		return mapModelPersistError(err)
	}
	return nil
}

func (s *service) ListRoutingRules(ctx context.Context) ([]types.RoutingRule, error) {
	rules, err := common.LoadRoutingRules(ctx, s.store.Org().Nodes(), s.store.Models().Allowlist())
	if err != nil {
		return nil, err
	}
	catalog, err := s.store.Models().Models(ctx)
	if err != nil {
		return nil, err
	}
	for i := range rules {
		rules[i] = common.EnrichRoutingRule(rules[i], catalog)
	}
	return rules, nil
}

func (s *service) ResolveRouting(ctx context.Context, deptID uuid.UUID) (types.ResolvedWhitelist, error) {
	departments, err := common.LoadDepartments(ctx, s.store.Org().Nodes())
	if err != nil {
		return types.ResolvedWhitelist{}, err
	}
	rules, err := common.LoadRoutingRules(ctx, s.store.Org().Nodes(), s.store.Models().Allowlist())
	if err != nil {
		return types.ResolvedWhitelist{}, err
	}
	models, err := s.store.Models().Models(ctx)
	if err != nil {
		return types.ResolvedWhitelist{}, err
	}
	rule := common.GetRoutingRuleForDept(deptID, rules, departments)
	if rule == nil {
		allowedIDs := modelcatalog.EnabledModelIDs(models)
		return types.ResolvedWhitelist{
			Inherited:       false,
			AllowedModelIDs: allowedIDs,
			AllowedModels:   modelcatalog.EnrichRefs(models, allowedIDs),
			ParentCount:     len(models),
		}, nil
	}
	parentID := common.GetParentDeptID(rule.NodeID, departments)
	parentCount := len(rule.AllowedModelIDs)
	if parentID != nil {
		for i := range rules {
			if rules[i].NodeID == *parentID {
				parentCount = len(rules[i].AllowedModelIDs)
				break
			}
		}
	}
	allowedModelIDs := common.ResolveDeptAllowedModelIDs(deptID, departments, rules, models)
	return types.ResolvedWhitelist{
		Inherited:       rule.Inherited,
		AllowedModelIDs: allowedModelIDs,
		AllowedModels:   modelcatalog.EnrichRefs(models, allowedModelIDs),
		ParentCount:     parentCount,
	}, nil
}

func (s *service) UpdateRoutingRule(
	ctx context.Context,
	id uuid.UUID,
	input types.UpdateRoutingRuleInput,
) (types.RoutingRule, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.RoutingRule{}, err
	}
	rules, err := common.LoadRoutingRules(ctx, s.store.Org().Nodes(), s.store.Models().Allowlist())
	if err != nil {
		return types.RoutingRule{}, err
	}
	idx := -1
	for i := range rules {
		if rules[i].ID == id {
			idx = i
			break
		}
	}
	if idx < 0 {
		return types.RoutingRule{}, domain.NotFound("Not found")
	}
	updated := rules[idx]
	if input.AllowedModelIDs != nil {
		if err := s.validateWritableModelIDs(ctx, input.AllowedModelIDs); err != nil {
			return types.RoutingRule{}, err
		}
		updated.AllowedModelIDs = append([]uuid.UUID{}, input.AllowedModelIDs...)
	}
	if input.Inherited != nil {
		updated.Inherited = *input.Inherited
	}
	if input.DefaultModelID != nil {
		if err := s.validateWritableModelIDs(ctx, []uuid.UUID{*input.DefaultModelID}); err != nil {
			return types.RoutingRule{}, err
		}
		updated.DefaultModelID = input.DefaultModelID
	}
	if input.FallbackModelID != nil {
		if err := s.validateWritableModelIDs(ctx, []uuid.UUID{*input.FallbackModelID}); err != nil {
			return types.RoutingRule{}, err
		}
		updated.FallbackModelID = input.FallbackModelID
	}
	rules[idx] = updated
	if input.AllowedModelIDs != nil {
		departments, err := common.LoadDepartments(ctx, s.store.Org().Nodes())
		if err != nil {
			return types.RoutingRule{}, err
		}
		rules = common.ShrinkChildRoutingRules(
			updated.NodeID,
			updated.AllowedModelIDs,
			rules,
			departments,
		)
	}
	nodes, err := s.store.Org().Nodes().Tree(ctx)
	if err != nil {
		return types.RoutingRule{}, err
	}
	if err := common.PersistRoutingRules(ctx, s.store, nodes, rules); err != nil {
		return types.RoutingRule{}, fmt.Errorf("persist routing rules: %w", err)
	}
	if s.client == nil {
		return types.RoutingRule{}, fmt.Errorf("newapi admin client required")
	}
	if err := s.client.RebuildAbilities(ctx); err != nil {
		return types.RoutingRule{}, fmt.Errorf("rebuild abilities: %w", err)
	}
	s.cacheInvalidator.InvalidateCompany(company.CompanyID(ctx))
	catalog, err := s.store.Models().Models(ctx)
	if err != nil {
		return types.RoutingRule{}, err
	}
	return common.EnrichRoutingRule(updated, catalog), nil
}

func (s *service) ListModelsWithPricing(ctx context.Context) ([]types.ModelInfoWithPricing, error) {
	models, err := s.store.Models().Models(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]types.ModelInfoWithPricing, len(models))
	for i := range models {
		result[i] = types.ModelInfoWithPricing{ModelInfo: models[i]}
	}

	// best-effort: NewAPI 不可达时仍返回模型列表，价格为 0
	if s.client == nil {
		return result, nil
	}
	pricingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	pricing, _ := s.client.ListModelPricing(pricingCtx)
	priceMap := make(map[string]adminport.ModelPricing, len(pricing))
	for _, p := range pricing {
		priceMap[p.ModelName] = p
	}

	for i := range result {
		if p, ok := priceMap[result[i].Type]; ok {
			result[i].InputPrice, result[i].OutputPrice = newapiunits.PriceFromRatio(p.ModelRatio, p.CompletionRatio)
		}
	}
	return result, nil
}
