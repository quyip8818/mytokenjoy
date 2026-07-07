package models

import (
	"context"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	ListModels(ctx context.Context) ([]types.ModelInfo, error)
	CreateModel(ctx context.Context, input types.CreateModelInput) (types.ModelInfo, error)
	UpdateModel(ctx context.Context, id string, input types.UpdateModelInput) (types.ModelInfo, error)
	DeleteModel(ctx context.Context, id string) error
	ToggleModel(ctx context.Context, id string, enabled bool) error
	ListRoutingRules(ctx context.Context) ([]types.RoutingRule, error)
	ResolveRouting(ctx context.Context, deptID string) (types.ResolvedWhitelist, error)
	UpdateRoutingRule(ctx context.Context, id string, input types.UpdateRoutingRuleInput) (types.RoutingRule, error)
}

type service struct {
	cfg         config.Config
	store       store.Store
	delayer     common.Delayer
	client      newapi.AdminClient
	modelLimits relay.ModelLimitsEnqueuer
}

func NewService(cfg config.Config, st store.Store, client newapi.AdminClient, modelLimits relay.ModelLimitsEnqueuer, delayer common.Delayer) Service {
	return &service{
		cfg:         cfg,
		store:       st,
		delayer:     delayer,
		client:      client,
		modelLimits: modelLimits,
	}
}

func (s *service) ListModels(ctx context.Context) ([]types.ModelInfo, error) {
	return s.store.Models().Models(ctx)
}

func (s *service) CreateModel(ctx context.Context, input types.CreateModelInput) (types.ModelInfo, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.ModelInfo{}, err
	}
	displayName := input.DisplayName
	if displayName == "" {
		displayName = input.Name
	}
	model := types.ModelInfo{
		ID:           fmt.Sprintf("model-%d", time.Now().UnixMilli()),
		Provider:     "custom",
		Name:         input.Name,
		DisplayName:  displayName,
		Type:         "custom",
		Description:  "",
		Visibility:   "all",
		Endpoint:     &input.BaseURL,
		InputPrice:   input.InputPrice,
		OutputPrice:  input.OutputPrice,
		MaxContext:   128000,
		Enabled:      true,
		Capabilities: []string{"chat"},
	}
	models, err := s.store.Models().Models(ctx)
	if err != nil {
		return types.ModelInfo{}, err
	}
	models = append(models, model)
	if err := s.store.Models().SetModels(ctx, models); err != nil {
		return types.ModelInfo{}, fmt.Errorf("persist models: %w", err)
	}
	return model, nil
}

func (s *service) UpdateModel(ctx context.Context, id string, input types.UpdateModelInput) (types.ModelInfo, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.ModelInfo{}, err
	}
	models, err := s.store.Models().Models(ctx)
	if err != nil {
		return types.ModelInfo{}, err
	}
	for i := range models {
		if models[i].ID != id {
			continue
		}
		if input.DisplayName != nil {
			models[i].DisplayName = *input.DisplayName
		}
		if input.Name != nil {
			models[i].Name = *input.Name
		}
		if input.Description != nil {
			models[i].Description = *input.Description
		}
		if input.Visibility != nil {
			models[i].Visibility = *input.Visibility
		}
		if input.Endpoint != nil && models[i].Type == "custom" {
			models[i].Endpoint = input.Endpoint
		}
		if input.InputPrice != nil {
			models[i].InputPrice = *input.InputPrice
		}
		if input.OutputPrice != nil {
			models[i].OutputPrice = *input.OutputPrice
		}
		if input.MaxContext != nil {
			models[i].MaxContext = *input.MaxContext
		}
		if input.Capabilities != nil {
			models[i].Capabilities = append([]string{}, input.Capabilities...)
		}
		if err := s.store.Models().SetModels(ctx, models); err != nil {
			return types.ModelInfo{}, fmt.Errorf("persist models: %w", err)
		}
		return models[i], nil
	}
	return types.ModelInfo{}, domain.NotFound("Not found")
}

func (s *service) DeleteModel(ctx context.Context, id string) error {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return err
	}
	models, err := s.store.Models().Models(ctx)
	if err != nil {
		return err
	}
	next := make([]types.ModelInfo, 0, len(models))
	found := false
	for _, model := range models {
		if model.ID == id {
			found = true
			continue
		}
		next = append(next, model)
	}
	if !found {
		return domain.NotFound("Not found")
	}
	if err := s.store.Models().SetModels(ctx, next); err != nil {
		return fmt.Errorf("persist models: %w", err)
	}
	return nil
}

func (s *service) ToggleModel(ctx context.Context, id string, enabled bool) error {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return err
	}
	models, err := s.store.Models().Models(ctx)
	if err != nil {
		return err
	}
	for i := range models {
		if models[i].ID == id {
			models[i].Enabled = enabled
			if err := s.store.Models().SetModels(ctx, models); err != nil {
				return fmt.Errorf("persist models: %w", err)
			}
			return nil
		}
	}
	return domain.NotFound("Not found")
}

func (s *service) ListRoutingRules(ctx context.Context) ([]types.RoutingRule, error) {
	return common.LoadRoutingRules(ctx, s.store.Org().Nodes(), s.store.Models().Allowlist())
}

func (s *service) ResolveRouting(ctx context.Context, deptID string) (types.ResolvedWhitelist, error) {
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
		allowed := make([]string, 0)
		for _, model := range models {
			if model.Enabled {
				allowed = append(allowed, model.Name)
			}
		}
		return types.ResolvedWhitelist{
			Inherited:     false,
			AllowedModels: allowed,
			ParentCount:   len(models),
		}, nil
	}
	parentID := common.GetParentDeptID(rule.NodeID, departments)
	parentCount := len(rule.AllowedModels)
	if parentID != nil {
		for i := range rules {
			if rules[i].NodeID == *parentID {
				parentCount = len(rules[i].AllowedModels)
				break
			}
		}
	}
	allowedModels := common.ResolveDeptAllowedModels(deptID, departments, rules, models)
	return types.ResolvedWhitelist{
		Inherited:     rule.Inherited,
		AllowedModels: allowedModels,
		ParentCount:   parentCount,
	}, nil
}

func (s *service) UpdateRoutingRule(
	ctx context.Context,
	id string,
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
	if input.AllowedModels != nil {
		updated.AllowedModels = append([]string{}, input.AllowedModels...)
	}
	if input.Inherited != nil {
		updated.Inherited = *input.Inherited
	}
	if input.DefaultModel != nil {
		updated.DefaultModel = input.DefaultModel
	}
	if input.FallbackModel != nil {
		updated.FallbackModel = input.FallbackModel
	}
	rules[idx] = updated
	if input.AllowedModels != nil {
		departments, err := common.LoadDepartments(ctx, s.store.Org().Nodes())
		if err != nil {
			return types.RoutingRule{}, err
		}
		rules = common.ShrinkChildRoutingRules(
			updated.NodeID,
			updated.AllowedModels,
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
	if s.client != nil && s.cfg.NewAPIEnabled {
		if err := s.client.RebuildAbilities(ctx); err != nil {
			return types.RoutingRule{}, fmt.Errorf("rebuild abilities: %w", err)
		}
		if s.modelLimits != nil {
			departments, err := common.LoadDepartments(ctx, s.store.Org().Nodes())
			if err != nil {
				return types.RoutingRule{}, err
			}
			deptIDs := org.CollectDescendantDeptIDs(departments, updated.NodeID)
			if err := s.modelLimits.EnqueueModelLimitsForDepartments(ctx, deptIDs); err != nil {
				return types.RoutingRule{}, fmt.Errorf("enqueue model limits: %w", err)
			}
		}
	}
	return updated, nil
}
