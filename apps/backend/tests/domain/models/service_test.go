package models_test

import (
	"context"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/models"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/tests/testutil"
)

func newModelsService(t *testing.T) models.Service {
	t.Helper()
	cfg, st := testutil.NewMemoryStoreFromConfig(t)
	return models.NewService(cfg, st, nil, nil)
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

func TestUpdateRoutingRuleShrinksChildren(t *testing.T) {
	svc := newModelsService(t)
	ctx := context.Background()
	if _, err := svc.UpdateRoutingRule(ctx, "rr-1", types.UpdateRoutingRuleInput{
		AllowedModels: []string{"gpt-4o"},
	}); err != nil {
		t.Fatal(err)
	}
	for _, rr := range svc.ListRoutingRules() {
		if rr.ID != "rr-2" && rr.ID != "rr-3" {
			continue
		}
		if len(rr.AllowedModels) != 1 || rr.AllowedModels[0] != "gpt-4o" {
			t.Fatalf("expected %s shrunk to [gpt-4o], got %v", rr.ID, rr.AllowedModels)
		}
	}
}

func TestResolveRoutingInherited(t *testing.T) {
	svc := newModelsService(t)
	resolved := svc.ResolveRouting("dept-3")
	if !resolved.Inherited {
		t.Fatal("expected dept-3 routing to be inherited")
	}
	if len(resolved.AllowedModels) == 0 {
		t.Fatal("expected non-empty allowed models for dept-3")
	}
}

func TestCreateModel(t *testing.T) {
	svc := newModelsService(t)
	created, err := svc.CreateModel(context.Background(), types.CreateModelInput{
		Name: "test-model", InputPrice: 1.0, OutputPrice: 2.0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Name != "test-model" || !created.Enabled {
		t.Fatalf("unexpected model %+v", created)
	}
	found := false
	for _, m := range svc.ListModels() {
		if m.ID == created.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("created model not in list")
	}
}

func TestToggleModel(t *testing.T) {
	svc := newModelsService(t)
	models := svc.ListModels()
	if len(models) == 0 {
		t.Fatal("expected models in seed")
	}
	target := models[0]
	wasEnabled := target.Enabled
	if err := svc.ToggleModel(context.Background(), target.ID, !wasEnabled); err != nil {
		t.Fatal(err)
	}
	for _, m := range svc.ListModels() {
		if m.ID == target.ID && m.Enabled == wasEnabled {
			t.Fatalf("expected enabled=%v, still %v", !wasEnabled, m.Enabled)
		}
	}
}
