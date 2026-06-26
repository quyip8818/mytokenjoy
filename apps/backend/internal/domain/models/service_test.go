package models_test

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/models"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/seed"
	"github.com/tokenjoy/backend/internal/store"
)

func newModelsService(t *testing.T) models.Service {
	t.Helper()
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	cfg.SimulateDelay = false
	return models.NewService(cfg, store.NewMemory(seed.Load(cfg)))
}

func TestResolveRoutingWithoutRule(t *testing.T) {
	svc := newModelsService(t)
	resolved := svc.ResolveRouting("missing-dept")
	if resolved.Inherited {
		t.Fatal("expected inherited false when no routing rule")
	}
	if len(resolved.AllowedModels) == 0 {
		t.Fatal("expected enabled models fallback")
	}
}

func TestUpdateRoutingRuleNotFound(t *testing.T) {
	svc := newModelsService(t)
	_, err := svc.UpdateRoutingRule(context.Background(), "missing", types.UpdateRoutingRuleInput{
		AllowedModels: []string{"gpt-4o"},
	})
	if err == nil {
		t.Fatal("expected not found error")
	}
}
