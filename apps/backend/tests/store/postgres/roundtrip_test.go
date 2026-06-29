//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/store/seed"
)

func TestBudgetTreeRoundTrip(t *testing.T) {
	st := testPostgresStore(t)
	parentID := "dept-roundtrip-parent"
	childID := "dept-roundtrip-child"
	parent := parentID
	tree := []types.BudgetNode{
		{
			ID:       parentID,
			Name:     "RoundTrip Root",
			Budget:   1000,
			Consumed: 0,
			Period:   "2026-06",
			Children: []types.BudgetNode{
				{
					ID:       childID,
					Name:     "RoundTrip Child",
					ParentID: &parent,
					Budget:   500,
					Consumed: 100,
					Period:   "2026-06",
				},
			},
		},
	}
	if err := st.Budget().SetTree(tree); err != nil {
		t.Fatal(err)
	}
	got := st.Budget().Tree()
	if len(got) != 1 || got[0].ID != parentID {
		t.Fatalf("unexpected root: %+v", got)
	}
	if len(got[0].Children) != 1 || got[0].Children[0].ID != childID {
		t.Fatalf("expected nested child, got %+v", got[0].Children)
	}
	if got[0].Children[0].Budget != 500 || got[0].Children[0].Consumed != 100 {
		t.Fatalf("child budget mismatch: %+v", got[0].Children[0])
	}
}

func TestKeysRoundTrip(t *testing.T) {
	st := testPostgresStore(t)
	keys := []types.ProviderKey{
		{
			ID:        "pk-roundtrip",
			Provider:  "openai",
			Name:      "RoundTrip Key",
			KeyPrefix: "sk-rt",
			SecretKey: "secret",
			Status:    "active",
			CreatedAt: "2026-06-01",
		},
	}
	if err := st.Keys().SetProviderKeys(keys); err != nil {
		t.Fatal(err)
	}
	got := st.Keys().ProviderKeys()
	found := false
	for _, key := range got {
		if key.ID == "pk-roundtrip" {
			found = true
			if key.Name != "RoundTrip Key" || key.Provider != "openai" {
				t.Fatalf("unexpected key: %+v", key)
			}
		}
	}
	if !found {
		t.Fatal("provider key not found after round-trip")
	}
}

func TestModelsRoutingRoundTrip(t *testing.T) {
	st := testPostgresStore(t)
	models := []types.ModelInfo{
		{
			ID:           "model-roundtrip",
			Provider:     "openai",
			Name:         "gpt-roundtrip",
			DisplayName:  "GPT RoundTrip",
			InputPrice:   1,
			OutputPrice:  2,
			MaxContext:   128000,
			Enabled:      true,
			Capabilities: []string{"chat"},
		},
	}
	defaultModel := "gpt-roundtrip"
	fallbackModel := "gpt-4o"
	rules := []types.RoutingRule{
		{
			ID:            "rr-roundtrip",
			NodeID:        seed.IDDept3,
			NodeName:      "后端组",
			DefaultModel:  &defaultModel,
			FallbackModel: &fallbackModel,
			AllowedModels: []string{"gpt-roundtrip", "gpt-4o"},
			Inherited:     false,
		},
	}
	if err := st.Models().SetModels(models); err != nil {
		t.Fatal(err)
	}
	if err := st.Models().SetRoutingRules(rules); err != nil {
		t.Fatal(err)
	}
	gotModels := st.Models().Models()
	foundModel := false
	for _, model := range gotModels {
		if model.ID == "model-roundtrip" {
			foundModel = true
			if len(model.Capabilities) != 1 || model.Capabilities[0] != "chat" {
				t.Fatalf("capabilities mismatch: %+v", model.Capabilities)
			}
		}
	}
	if !foundModel {
		t.Fatal("model not found after round-trip")
	}
	gotRules := st.Models().RoutingRules()
	foundRule := false
	for _, rule := range gotRules {
		if rule.ID == "rr-roundtrip" {
			foundRule = true
			if len(rule.AllowedModels) != 2 || rule.DefaultModel == nil || *rule.DefaultModel != "gpt-roundtrip" {
				t.Fatalf("routing rule mismatch: %+v", rule)
			}
		}
	}
	if !foundRule {
		t.Fatal("routing rule not found after round-trip")
	}
}

func TestWithTxRollback(t *testing.T) {
	st := testPostgresStore(t)
	pool := testDBPool(t)
	ctx := context.Background()

	before := memberUpdatedAt(t, pool, seed.IDMember1)
	originalName := findMemberName(st.Org().Members(), seed.IDMember1)
	if originalName == "" {
		t.Fatalf("member %s not found", seed.IDMember1)
	}

	err := st.WithTx(ctx, func(tx store.Store) error {
		members := tx.Org().Members()
		for i := range members {
			if members[i].ID == seed.IDMember1 {
				members[i].Name = "ShouldRollback"
			}
		}
		if err := tx.Org().SetMembers(members); err != nil {
			return err
		}
		return errors.New("force rollback")
	})
	if err == nil {
		t.Fatal("expected transaction error")
	}

	after := memberUpdatedAt(t, pool, seed.IDMember1)
	if got := findMemberName(st.Org().Members(), seed.IDMember1); got != originalName {
		t.Fatalf("expected name %q after rollback, got %q", originalName, got)
	}
	if !after.Equal(before) {
		t.Fatalf("expected member updated_at unchanged on rollback: before=%v after=%v", before, after)
	}
}
