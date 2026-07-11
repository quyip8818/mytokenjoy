package models_test

import (
	"strconv"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/models"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func newModelsService(t *testing.T) models.Service {
	t.Helper()
	cfg, st := testutil.NewTestStore(t)
	return models.NewService(cfg, st, newapi.NewAdminPortAdapter(&mock.StubAdminClient{}), nil, common.NewDelayer(false))
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
		AllowedModelIDs: []int64{contract.IDModel1},
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
		AllowedModelIDs: []int64{contract.IDModel1},
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
		if len(rr.AllowedModelIDs) != 1 || rr.AllowedModelIDs[0] != contract.IDModel1 {
			t.Fatalf("expected %s shrunk to [gpt-4o], got %v", rr.ID, rr.AllowedModelIDs)
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
		Type: "test-model", InputPrice: 1.0, OutputPrice: 2.0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Type != "test-model" || !created.Enabled {
		t.Fatalf("unexpected model %+v", created)
	}
	found := false
	models, err := svc.ListModels(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range models {
		if m.ModelID == created.ModelID {
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
	target, err := svc.CreateModel(testutil.Ctx(), types.CreateModelInput{
		Type: "toggle-me", InputPrice: 1.0, OutputPrice: 2.0,
	})
	if err != nil {
		t.Fatal(err)
	}
	wasEnabled := target.Enabled
	if err := svc.ToggleModel(testutil.Ctx(), strconv.FormatInt(target.ModelID, 10), !wasEnabled); err != nil {
		t.Fatal(err)
	}
	after, err := svc.ListModels(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range after {
		if m.ModelID == target.ModelID && m.Enabled == wasEnabled {
			t.Fatalf("expected enabled=%v, still %v", !wasEnabled, m.Enabled)
		}
	}
}

func TestToggleGlobalModelRejected(t *testing.T) {
	t.Parallel()
	svc := newModelsService(t)
	models, err := svc.ListModels(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if len(models) == 0 {
		t.Fatal("expected models in seed")
	}
	var global *types.ModelInfo
	for i := range models {
		if models[i].Provider != types.ProviderCustom {
			global = &models[i]
			break
		}
	}
	if global == nil {
		t.Fatal("expected builtin model")
	}
	err = svc.ToggleModel(testutil.Ctx(), strconv.FormatInt(global.ModelID, 10), !global.Enabled)
	if err == nil {
		t.Fatal("expected global model toggle to be rejected")
	}
}

func TestUpdateModel(t *testing.T) {
	t.Parallel()
	svc := newModelsService(t)
	created, err := svc.CreateModel(testutil.Ctx(), types.CreateModelInput{
		Type: "update-me", InputPrice: 1.0, OutputPrice: 2.0,
	})
	if err != nil {
		t.Fatal(err)
	}
	label := "Updated Display"
	updated, err := svc.UpdateModel(testutil.Ctx(), strconv.FormatInt(created.ModelID, 10), types.UpdateModelInput{
		Name: &label,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != label {
		t.Fatalf("expected name %q, got %q", label, updated.Name)
	}
}

func TestCreateModelAllowsSameTypeDifferentProvider(t *testing.T) {
	t.Parallel()
	svc := newModelsService(t)
	created, err := svc.CreateModel(testutil.Ctx(), types.CreateModelInput{
		Type: "gpt-4o", InputPrice: 1.0, OutputPrice: 2.0, BaseURL: "http://llm.test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Provider != types.ProviderCustom || created.Type != "gpt-4o" {
		t.Fatalf("unexpected model %+v", created)
	}
	models, err := svc.ListModels(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	var builtinCount, customCount int
	for _, m := range models {
		if m.Type != "gpt-4o" {
			continue
		}
		if m.Provider == types.ProviderCustom {
			customCount++
		} else {
			builtinCount++
		}
	}
	if builtinCount == 0 || customCount == 0 {
		t.Fatalf("expected builtin and custom gpt-4o in catalog, got builtin=%d custom=%d", builtinCount, customCount)
	}
}

func TestCreateModelRejectsDuplicate(t *testing.T) {
	t.Parallel()
	svc := newModelsService(t)
	ctx := testutil.Ctx()
	if _, err := svc.CreateModel(ctx, types.CreateModelInput{
		Type: "dup-model", InputPrice: 1.0, OutputPrice: 2.0, BaseURL: "http://llm.test",
	}); err != nil {
		t.Fatal(err)
	}
	_, err := svc.CreateModel(ctx, types.CreateModelInput{
		Type: "dup-model", InputPrice: 1.0, OutputPrice: 2.0, BaseURL: "http://llm.test",
	})
	if err == nil {
		t.Fatal("expected error for duplicate custom model name")
	}
}

func TestDeleteModel(t *testing.T) {
	t.Parallel()
	svc := newModelsService(t)
	created, err := svc.CreateModel(testutil.Ctx(), types.CreateModelInput{
		Type: "delete-me", InputPrice: 1.0, OutputPrice: 2.0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.DeleteModel(testutil.Ctx(), strconv.FormatInt(created.ModelID, 10)); err != nil {
		t.Fatal(err)
	}
	models, err := svc.ListModels(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range models {
		if m.ModelID == created.ModelID {
			t.Fatal("deleted model still in list")
		}
	}
}
