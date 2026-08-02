//go:build testhook

package models_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/models"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
	"github.com/tokenjoy/backend/tests/testutil/mock"
)

// TestSyncFromPlatform_DeactivatesStaleModels verifies the global catalog sync:
// 1. Insert platform models → they appear in ListModels
// 2. Sync again with a smaller set → removed models become inactive (not in ListModels)
func TestSyncFromPlatform_DeactivatesStaleModels(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()
	repo := st.Models()

	globalID := cfg.TokenJoyCompanyID

	// Sync 3 models.
	err := repo.SyncFromPlatform(ctx, globalID, []types.ModelInfo{
		{Provider: "openai", Type: "gpt-4o", Name: "GPT-4o"},
		{Provider: "openai", Type: "gpt-4o-mini", Name: "GPT-4o Mini"},
		{Provider: "anthropic", Type: "claude-sonnet", Name: "Claude Sonnet"},
	})
	if err != nil {
		t.Fatalf("first sync: %v", err)
	}

	// All 3 should be visible via Models().
	all, err := repo.Models(ctx)
	if err != nil {
		t.Fatal(err)
	}
	platformCount := 0
	for _, m := range all {
		if m.Source == "platform" && !m.Deprecated {
			platformCount++
		}
	}
	if platformCount != 3 {
		t.Fatalf("expected 3 active platform models, got %d", platformCount)
	}

	// Sync again with only 2 models (claude-sonnet removed).
	err = repo.SyncFromPlatform(ctx, globalID, []types.ModelInfo{
		{Provider: "openai", Type: "gpt-4o", Name: "GPT-4o"},
		{Provider: "openai", Type: "gpt-4o-mini", Name: "GPT-4o Mini"},
	})
	if err != nil {
		t.Fatalf("second sync: %v", err)
	}

	// Now only 2 should be active; claude-sonnet should be inactive.
	all, err = repo.Models(ctx)
	if err != nil {
		t.Fatal(err)
	}
	activeCount := 0
	var inactiveModel *types.ModelInfo
	for i, m := range all {
		if m.Source != "platform" {
			continue
		}
		if !m.Deprecated {
			activeCount++
		} else {
			inactiveModel = &all[i]
		}
	}
	if activeCount != 2 {
		t.Fatalf("expected 2 active platform models after second sync, got %d", activeCount)
	}
	if inactiveModel == nil || inactiveModel.Type != "claude-sonnet" {
		t.Fatalf("expected claude-sonnet to be inactive, got %+v", inactiveModel)
	}
}

// TestToggleModel_RejectsEnablingGloballyInactiveModel verifies that UpdateModel
// rejects platform models (requireTenantModel check).
func TestToggleModel_RejectsEnablingGloballyInactiveModel(t *testing.T) {
	t.Parallel()
	cfg, st := testutil.NewTestStore(t)
	ctx := testutil.Ctx()
	svc := models.NewService(cfg, st, &mock.StubAdminClient{}, nil, common.NewDelayer(false))

	globalID := cfg.TokenJoyCompanyID

	// Insert a global platform model, then deactivate it (simulating platform delist).
	err := st.Models().SyncFromPlatform(ctx, globalID, []types.ModelInfo{
		{Provider: "openai", Type: "toggle-reject-test", Name: "Toggle Reject"},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Sync with empty list → deactivates it.
	err = st.Models().SyncFromPlatform(ctx, globalID, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Find the model via repo (it's inactive but still exists).
	model, err := st.Models().GlobalModelByProviderType(ctx, "openai", "toggle-reject-test")
	if err != nil || model == nil {
		t.Fatalf("expected to find global model: err=%v model=%v", err, model)
	}
	if !model.Deprecated {
		t.Fatal("expected model to be inactive after empty sync")
	}

	// UpdateModel on a platform model should be rejected (requireTenantModel).
	trueVal := true
	_, err = svc.UpdateModel(ctx, model.ID, types.UpdateModelInput{Deprecated: &trueVal})
	if err == nil {
		t.Fatal("expected UpdateModel to reject platform model")
	}
}

// TestListModels_ExcludesInactiveButRoutingStillEnriches verifies that:
// - ListModels does NOT return inactive models (display layer)
// - ListRoutingRules still shows model names for inactive models in existing configs
func TestListModels_ExcludesInactiveButRoutingStillEnriches(t *testing.T) {
	t.Parallel()
	svc := newModelsService(t)
	ctx := testutil.Ctx()

	// Seed has IDModel1 (deepseek-v4-pro) as active, owned by DefaultCompanyID.
	// It's also in org_node allowlists. Verify it appears in ListModels.
	before, err := svc.ListModels(ctx)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range before {
		if m.ID == contract.IDModel1 {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected IDModel1 in ListModels before deactivation")
	}

	// Deactivate IDModel1 directly (simulate SMS delist for tenant-owned model).
	model1, err := svc.ListModels(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var target *types.ModelInfo
	for i := range model1 {
		if model1[i].ID == contract.IDModel1 {
			target = &model1[i]
			break
		}
	}
	if target == nil {
		t.Fatal("IDModel1 not found")
	}

	// Toggle it off via UpdateModel (target is a custom/tenant model).
	trueVal := true
	if _, err := svc.UpdateModel(ctx, target.ID, types.UpdateModelInput{Deprecated: &trueVal}); err != nil {
		t.Fatal(err)
	}

	// ListModels should NOT include it anymore.
	after, err := svc.ListModels(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range after {
		if m.ID == contract.IDModel1 {
			t.Fatal("inactive model should not appear in ListModels")
		}
	}

	// But ListRoutingRules should still enrich its name in existing routing rules.
	rules, err := svc.ListRoutingRules(ctx)
	if err != nil {
		t.Fatal(err)
	}
	foundInRouting := false
	for _, rule := range rules {
		for _, ref := range rule.AllowedModels {
			if ref.ID == contract.IDModel1 {
				foundInRouting = true
				if ref.Name == "" {
					t.Fatal("expected enriched name for inactive model in routing rule")
				}
				break
			}
		}
		if foundInRouting {
			break
		}
	}
	if !foundInRouting {
		t.Fatal("expected IDModel1 to still appear in routing rules after deactivation")
	}
}
