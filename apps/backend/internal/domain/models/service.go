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
	"github.com/tokenjoy/backend/internal/pkg/queryutil"
	"github.com/tokenjoy/backend/internal/pkg/routingutil"
	"github.com/tokenjoy/backend/internal/pkg/simulate"
	"github.com/tokenjoy/backend/internal/store"
)

type Service interface {
	ListModels() []types.ModelInfo
	CreateModel(ctx context.Context, input types.CreateModelInput) (types.ModelInfo, error)
	ToggleModel(ctx context.Context, id string, enabled bool) error
	ListRoutingRules() []types.RoutingRule
	ResolveRouting(deptID string) types.ResolvedWhitelist
	UpdateRoutingRule(ctx context.Context, id string, input types.UpdateRoutingRuleInput) (types.RoutingRule, error)
}

type service struct {
	cfg       config.Config
	store     store.Store
	delayer   simulate.Delayer
	client    newapi.AdminClient
	lifecycle relay.Lifecycle
}

func NewService(cfg config.Config, st store.Store, client newapi.AdminClient, lifecycle relay.Lifecycle, delayer simulate.Delayer) Service {
	return &service{
		cfg:       cfg,
		store:     st,
		delayer:   delayer,
		client:    client,
		lifecycle: lifecycle,
	}
}

func (s *service) ListModels() []types.ModelInfo {
	return s.store.Models().Models()
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
		InputPrice:   input.InputPrice,
		OutputPrice:  input.OutputPrice,
		MaxContext:   128000,
		Enabled:      true,
		Capabilities: []string{"chat"},
	}
	models := s.store.Models().Models()
	models = append(models, model)
	if err := s.store.Models().SetModels(models); err != nil {
		return types.ModelInfo{}, fmt.Errorf("persist models: %w", err)
	}
	return model, nil
}

func (s *service) ToggleModel(ctx context.Context, id string, enabled bool) error {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return err
	}
	models := s.store.Models().Models()
	for i := range models {
		if models[i].ID == id {
			models[i].Enabled = enabled
			if err := s.store.Models().SetModels(models); err != nil {
				return fmt.Errorf("persist models: %w", err)
			}
			return nil
		}
	}
	return domain.NotFound("Not found")
}

func (s *service) ListRoutingRules() []types.RoutingRule {
	return s.store.Models().RoutingRules()
}

func (s *service) ResolveRouting(deptID string) types.ResolvedWhitelist {
	departments := s.store.Org().Departments()
	rules := s.store.Models().RoutingRules()
	models := s.store.Models().Models()
	rule := routingutil.GetRoutingRuleForDept(deptID, rules, departments)
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
		}
	}
	parentID := routingutil.GetParentDeptID(rule.NodeID, departments)
	parentCount := len(rule.AllowedModels)
	if parentID != nil {
		for i := range rules {
			if rules[i].NodeID == *parentID {
				parentCount = len(rules[i].AllowedModels)
				break
			}
		}
	}
	allowedModels := routingutil.ResolveDeptAllowedModels(deptID, departments, rules, models)
	return types.ResolvedWhitelist{
		Inherited:     rule.Inherited,
		AllowedModels: allowedModels,
		ParentCount:   parentCount,
	}
}

func (s *service) UpdateRoutingRule(
	ctx context.Context,
	id string,
	input types.UpdateRoutingRuleInput,
) (types.RoutingRule, error) {
	if err := s.delayer.Wait(ctx, 300*time.Millisecond); err != nil {
		return types.RoutingRule{}, err
	}
	rules := s.store.Models().RoutingRules()
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
		rules = routingutil.ShrinkChildRoutingRules(
			updated.NodeID,
			updated.AllowedModels,
			rules,
			s.store.Org().Departments(),
		)
	}
	if err := s.store.Models().SetRoutingRules(rules); err != nil {
		return types.RoutingRule{}, fmt.Errorf("persist routing rules: %w", err)
	}
	if s.client != nil && s.cfg.NewAPIEnabled {
		if err := s.client.RebuildAbilities(ctx); err != nil {
			return types.RoutingRule{}, fmt.Errorf("rebuild abilities: %w", err)
		}
		if s.lifecycle != nil {
			deptIDs := queryutil.CollectDescendantDeptIDs(s.store.Org().Departments(), updated.NodeID)
			if err := s.lifecycle.EnqueueModelLimitsForDepartments(deptIDs); err != nil {
				return types.RoutingRule{}, fmt.Errorf("enqueue model limits: %w", err)
			}
		}
	}
	return updated, nil
}
