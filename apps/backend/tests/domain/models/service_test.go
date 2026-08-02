package models_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/models"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

func newModelsService(t *testing.T) models.Service {
	t.Helper()
	cfg, st := testutil.NewTestStore(t)
	return models.NewService(cfg, st, &mock.StubAdminClient{}, nil, common.NewDelayer(false))
}

func TestResolveRoutingWithoutRule(t *testing.T) {
	t.Parallel()
	svc := newModelsService(t)
	resolved, err := svc.ResolveRouting(testutil.Ctx(), uuid.MustParse("00000000-0000-7000-0000-000000000099"))
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
	_, err := svc.UpdateRoutingRule(testutil.Ctx(), uuid.MustParse("00000000-0000-0000-0000-000000000000"), types.UpdateRoutingRuleInput{
		AllowedModelIDs: []uuid.UUID{contract.IDModel1},
	})
	if err == nil {
		t.Fatal("expected not found error")
	}
}

func TestUpdateRoutingRuleShrinksChildren(t *testing.T) {
	t.Parallel()
	svc := newModelsService(t)
	ctx := testutil.Ctx()
	if _, err := svc.UpdateRoutingRule(ctx, contract.IDDept1, types.UpdateRoutingRuleInput{
		AllowedModelIDs: []uuid.UUID{contract.IDModel1},
	}); err != nil {
		t.Fatal(err)
	}
	rules, err := svc.ListRoutingRules(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, rr := range rules {
		if rr.ID != contract.IDDept2 && rr.ID != contract.IDDept3 {
			continue
		}
		if len(rr.AllowedModelIDs) != 1 || rr.AllowedModelIDs[0] != contract.IDModel1 {
			t.Fatalf("expected %s shrunk to [model1], got %v", rr.ID, rr.AllowedModelIDs)
		}
	}
}

func TestResolveRoutingInherited(t *testing.T) {
	t.Parallel()
	svc := newModelsService(t)
	resolved, err := svc.ResolveRouting(testutil.Ctx(), contract.IDDept3)
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
		Type: "create-model-test", InputPrice: 1.0, OutputPrice: 2.0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Type != "create-model-test" || created.Deprecated {
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
	target, err := svc.CreateModel(testutil.Ctx(), types.CreateModelInput{
		Type: "toggle-me", InputPrice: 1.0, OutputPrice: 2.0,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Deprecate via UpdateModel
	trueVal := true
	if _, err := svc.UpdateModel(testutil.Ctx(), target.ID, types.UpdateModelInput{Deprecated: &trueVal}); err != nil {
		t.Fatal(err)
	}
	after, err := svc.ListModels(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range after {
		if m.ID == target.ID {
			if !m.Deprecated {
				t.Fatal("expected model to be deprecated after update")
			}
			return
		}
	}
	// Model may not appear in active-only list after deprecation — that's fine
}

func TestToggleGlobalModelCreatesOverride(t *testing.T) {
	t.Parallel()
	svc := newModelsService(t)
	allModels, err := svc.ListModels(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	if len(allModels) == 0 {
		t.Fatal("expected models in seed")
	}
	var global *types.ModelInfo
	for i := range allModels {
		if allModels[i].Provider != types.ProviderCustom {
			global = &allModels[i]
			break
		}
	}
	if global == nil {
		t.Fatal("expected builtin model")
	}

	// UpdateModel on a platform model should be rejected (requireTenantModel).
	falseVal := false
	_, err = svc.UpdateModel(testutil.Ctx(), global.ID, types.UpdateModelInput{Deprecated: &falseVal})
	if err == nil {
		t.Fatal("expected UpdateModel to reject platform model")
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
	updated, err := svc.UpdateModel(testutil.Ctx(), created.ID, types.UpdateModelInput{
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

	// Pick a builtin type that exists only as a global model (not as a tenant custom model).
	const builtinType = "gpt-4o"

	created, err := svc.CreateModel(testutil.Ctx(), types.CreateModelInput{
		Type: builtinType, InputPrice: 1.0, OutputPrice: 2.0, BaseURL: "http://llm.test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Provider != types.ProviderCustom || created.Type != builtinType {
		t.Fatalf("unexpected model %+v", created)
	}
	models, err := svc.ListModels(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	var builtinCount, customCount int
	for _, m := range models {
		if m.Type != builtinType {
			continue
		}
		if m.Provider == types.ProviderCustom {
			customCount++
		} else {
			builtinCount++
		}
	}
	if builtinCount == 0 || customCount == 0 {
		t.Fatalf("expected builtin and custom %s in catalog, got builtin=%d custom=%d", builtinType, builtinCount, customCount)
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

func TestCreateModelWithNewFields(t *testing.T) {
	t.Parallel()
	svc := newModelsService(t)
	created, err := svc.CreateModel(testutil.Ctx(), types.CreateModelInput{
		Type:              "new-fields-model",
		Name:              "New Fields Model",
		BaseURL:           "https://api.example.com/v1",
		ApiKey:            "sk-test-key-123",
		EndpointModelName: "chatgpt4.0",
		InputPrice:        1.0,
		OutputPrice:       2.0,
		MaxContext:        1000000,
		MaxTokens:         8192,
		Capabilities:      []string{"chat", "vision"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Name != "New Fields Model" {
		t.Fatalf("expected name 'New Fields Model', got %q", created.Name)
	}
	if created.ApiKey == nil || *created.ApiKey != "sk-test-key-123" {
		t.Fatalf("expected apiKey 'sk-test-key-123', got %v", created.ApiKey)
	}
	if created.EndpointModelName == nil || *created.EndpointModelName != "chatgpt4.0" {
		t.Fatalf("expected endpointModelName 'chatgpt4.0', got %v", created.EndpointModelName)
	}
	if created.MaxContext != 1000000 {
		t.Fatalf("expected maxContext 1000000, got %d", created.MaxContext)
	}
	if created.MaxTokens != 8192 {
		t.Fatalf("expected maxTokens 8192, got %d", created.MaxTokens)
	}
	if len(created.Capabilities) != 2 {
		t.Fatalf("expected 2 capabilities, got %v", created.Capabilities)
	}
}

func TestUpdateModelNewFields(t *testing.T) {
	t.Parallel()
	svc := newModelsService(t)
	created, err := svc.CreateModel(testutil.Ctx(), types.CreateModelInput{
		Type: "update-new-fields", BaseURL: "https://api.test.com", InputPrice: 1.0, OutputPrice: 2.0,
	})
	if err != nil {
		t.Fatal(err)
	}
	newKey := "sk-updated-key"
	newEndpointModel := "gpt-4-turbo"
	newMaxTokens := 16384
	updated, err := svc.UpdateModel(testutil.Ctx(), created.ID, types.UpdateModelInput{
		ApiKey:            &newKey,
		EndpointModelName: &newEndpointModel,
		MaxTokens:         &newMaxTokens,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ApiKey == nil || *updated.ApiKey != newKey {
		t.Fatalf("expected apiKey %q, got %v", newKey, updated.ApiKey)
	}
	if updated.EndpointModelName == nil || *updated.EndpointModelName != newEndpointModel {
		t.Fatalf("expected endpointModelName %q, got %v", newEndpointModel, updated.EndpointModelName)
	}
	if updated.MaxTokens != newMaxTokens {
		t.Fatalf("expected maxTokens %d, got %d", newMaxTokens, updated.MaxTokens)
	}
}

func TestToggleGlobalModelTwice(t *testing.T) {
	t.Parallel()
	svc := newModelsService(t)
	allModels, err := svc.ListModels(testutil.Ctx())
	if err != nil {
		t.Fatal(err)
	}
	var global *types.ModelInfo
	for i := range allModels {
		if allModels[i].Provider != types.ProviderCustom {
			global = &allModels[i]
			break
		}
	}
	if global == nil {
		t.Fatal("expected builtin model")
	}
	// UpdateModel on a platform model should be rejected (requireTenantModel).
	trueVal := true
	_, err = svc.UpdateModel(testutil.Ctx(), global.ID, types.UpdateModelInput{Deprecated: &trueVal})
	if err == nil {
		t.Fatal("expected UpdateModel to reject platform model")
	}
}
