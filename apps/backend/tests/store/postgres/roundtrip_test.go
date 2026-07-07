package postgres_test

import (
	"errors"
	"testing"

	orgfix "github.com/tokenjoy/backend/tests/testutil/org"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestOrgNodesBudgetRoundTrip(t *testing.T) {
	t.Parallel()
	st := testPostgresStore(t)
	ctx := testutil.Ctx()
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
	nodes := orgfix.OrgNodesFromBudgetTree(tree)
	if err := st.Org().Nodes().SetTree(ctx, nodes); err != nil {
		t.Fatal(err)
	}
	got, err := common.LoadBudgetTree(ctx, st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
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
	t.Parallel()
	st := testPostgresStore(t)
	ctx := testutil.Ctx()
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
	if err := st.Keys().SetProviderKeys(ctx, keys); err != nil {
		t.Fatal(err)
	}
	got, err := st.Keys().ProviderKeys(ctx)
	if err != nil {
		t.Fatal(err)
	}
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

func TestModelAllowlistRoutingRoundTrip(t *testing.T) {
	t.Parallel()
	st := testPostgresStore(t)
	ctx := testutil.Ctx()
	models := []types.ModelInfo{
		{
			ID:           "model-roundtrip",
			Provider:     "openai",
			Name:         "gpt-roundtrip",
			DisplayName:  "GPT RoundTrip",
			Type:         "builtin",
			Visibility:   "all",
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
			ID:            contract.IDDept3,
			NodeID:        contract.IDDept3,
			NodeName:      "后端组",
			DefaultModel:  &defaultModel,
			FallbackModel: &fallbackModel,
			AllowedModels: []string{"gpt-roundtrip", "gpt-4o"},
			Inherited:     false,
		},
	}
	if err := st.Models().SetModels(ctx, models); err != nil {
		t.Fatal(err)
	}
	nodes, err := st.Org().Nodes().Tree(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := common.PersistRoutingRules(ctx, st, nodes, rules); err != nil {
		t.Fatal(err)
	}
	gotModels, err := st.Models().Models(ctx)
	if err != nil {
		t.Fatal(err)
	}
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
	gotRules, err := common.LoadRoutingRules(ctx, st.Org().Nodes(), st.Models().Allowlist())
	if err != nil {
		t.Fatal(err)
	}
	foundRule := false
	for _, rule := range gotRules {
		if rule.ID == contract.IDDept3 {
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
	t.Parallel()
	st := testPostgresStore(t)
	pool := testDBPool(t)
	ctx := testutil.Ctx()

	before := memberUpdatedAt(t, pool, contract.IDMember1)
	members, err := st.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	originalName := findMemberName(members, contract.IDMember1)
	if originalName == "" {
		t.Fatalf("member %s not found", contract.IDMember1)
	}

	err = st.WithTx(ctx, func(tx store.Store) error {
		members, err := tx.Org().Members(ctx)
		if err != nil {
			return err
		}
		for i := range members {
			if members[i].ID == contract.IDMember1 {
				members[i].Name = "ShouldRollback"
			}
		}
		if err := tx.Org().SetMembers(ctx, members); err != nil {
			return err
		}
		return errors.New("force rollback")
	})
	if err == nil {
		t.Fatal("expected transaction error")
	}

	after := memberUpdatedAt(t, pool, contract.IDMember1)
	members, err = st.Org().Members(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got := findMemberName(members, contract.IDMember1); got != originalName {
		t.Fatalf("expected name %q after rollback, got %q", originalName, got)
	}
	if !after.Equal(before) {
		t.Fatalf("expected member updated_at unchanged on rollback: before=%v after=%v", before, after)
	}
}
