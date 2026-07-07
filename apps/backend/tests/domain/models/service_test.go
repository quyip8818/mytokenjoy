package models_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/models"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/tests/testutil"
)

func newModelsService(t *testing.T) models.Service {
	t.Helper()
	cfg, st := testutil.NewTestStore(t)
	return models.NewService(cfg, st, nil, nil, common.NewDelayer(false))
}

func TestResolveRoutingWithoutRule(t *testing.T) {
	t.Parallel()
	svc := newModelsService(t)
	resolved, err := svc.ResolveRouting(testutil.Ctx(), "missing-dept")
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Inherited {
		t.Fatal("expected inherited false when no routing rule")
	}
	if len(resolved.AllowedModels) == 0 {
		t.Fatal("expected enabled models fallback")
	}
}

func TestUpdateRoutingRuleNotFound(t *testing.T) {
	t.Parallel()
	svc := newModelsService(t)
	_, err := svc.UpdateRoutingRule(testutil.Ctx(), "missing", types.UpdateRoutingRuleInput{
		AllowedModels: []string{"gpt-4o"},
	})
	if err == nil {
		t.Fatal("expected not found error")
	}
}

func TestUpdateRoutingRuleShrinksChildren(t *testing.T) {
	t.Parallel()
	svc := newModelsService(t)
	ctx := testutil.Ctx()
	if _, err := svc.UpdateRoutingRule(ctx, "dept-1", types.UpdateRoutingRuleInput{
		AllowedModels: []string{"gpt-4o"},
	}); err != nil {
		t.Fatal(err)
	}
	rules, err := svc.ListRoutingRules(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, rr := range rules {
		if rr.ID != "dept-2" && rr.ID != "dept-3" {
			continue
		}
		if len(rr.AllowedModels) != 1 || rr.AllowedModels[0] != "gpt-4o" {
			t.Fatalf("expected %s shrunk to [gpt-4o], got %v", rr.ID, rr.AllowedModels)
		}
	}
}

func TestResolveRoutingInherited(t *testing.T) {
	t.Parallel()
	svc := newModelsService(t)
	resolved, err := svc.ResolveRouting(testutil.Ctx(), "dept-3")
	if err != nil {
		t.Fatal(err)
	}
	if !resolved.Inherited {
		t.Fatal("expected dept-3 routing to be inherited")
	}
	if len(resolved.AllowedModels) == 0 {
		t.Fatal("expected non-empty allowed models for dept-3")
	}
}

func TestCreateModel(t *testing.T) {
	t.Parallel()
	svc := newModelsService(t)
	created, err := svc.CreateModel(testutil.Ctx(), types.CreateModelInput{
		Name: "test-model", InputPrice: 1.0, OutputPrice: 2.0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Name != "test-model" || !created.Enabled {
		t.Fatalf("unexpected model %+v", created)
	}
	found := false
	models, err := svc.ListModels(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range models {
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
	t.Parallel()
	svc := newModelsService(t)
	models, err := svc.ListModels(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if len(models) == 0 {
		t.Fatal("expected models in seed")
	}
	target := models[0]
	wasEnabled := target.Enabled
	if err := svc.ToggleModel(testutil.Ctx(), target.ID, !wasEnabled); err != nil {
		t.Fatal(err)
	}
	after, err := svc.ListModels(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range after {
		if m.ID == target.ID && m.Enabled == wasEnabled {
			t.Fatalf("expected enabled=%v, still %v", !wasEnabled, m.Enabled)
		}
	}
}

func TestUpdateModel(t *testing.T) {
	t.Parallel()
	svc := newModelsService(t)
	created, err := svc.CreateModel(testutil.Ctx(), types.CreateModelInput{
		Name: "update-me", InputPrice: 1.0, OutputPrice: 2.0,
	})
	if err != nil {
		t.Fatal(err)
	}
	displayName := "Updated Display"
	updated, err := svc.UpdateModel(testutil.Ctx(), created.ID, types.UpdateModelInput{
		DisplayName: &displayName,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.DisplayName != displayName {
		t.Fatalf("expected displayName %q, got %q", displayName, updated.DisplayName)
	}
}

func TestDeleteModel(t *testing.T) {
	t.Parallel()
	svc := newModelsService(t)
	created, err := svc.CreateModel(testutil.Ctx(), types.CreateModelInput{
		Name: "delete-me", InputPrice: 1.0, OutputPrice: 2.0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.DeleteModel(testutil.Ctx(), created.ID); err != nil {
		t.Fatal(err)
	}
	models, err := svc.ListModels(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range models {
		if m.ID == created.ID {
			t.Fatal("deleted model still in list")
		}
	}
}
