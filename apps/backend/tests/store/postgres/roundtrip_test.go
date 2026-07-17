package postgres_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	orgfix "github.com/tokenjoy/backend/tests/testutil/org"

	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestOrgNodesBudgetRoundTrip(t *testing.T) {
	t.Parallel()
	st := testPostgresStore(t)
	ctx := testutil.Ctx()
	tree, err := common.LoadBudgetTree(ctx, st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	child := pkgbudget.FindBudgetNode(tree, contract.IDDept3)
	if child == nil {
		t.Fatal("dept-3 not found")
	}
	child.Budget = 500
	if !pkgbudget.UpdateBudgetNodeInTree(tree, contract.IDDept3, types.BudgetNode{Budget: 500}) {
		t.Fatal("update budget node")
	}
	if err := orgfix.PersistBudgetTree(ctx, st, tree); err != nil {
		t.Fatal(err)
	}
	got, err := common.LoadBudgetTree(ctx, st.Org().Nodes())
	if err != nil {
		t.Fatal(err)
	}
	updated := pkgbudget.FindBudgetNode(got, contract.IDDept3)
	if updated == nil || updated.Budget != 500 {
		t.Fatalf("child budget mismatch: %+v", updated)
	}
}

func TestKeysRoundTrip(t *testing.T) {
	t.Parallel()
	st := testPostgresStore(t)
	ctx := testutil.Ctx()
	keys := []types.ProviderKey{
		{
			ID:        uuid.MustParse("00000000-0000-7000-0000-00000000ff77"),
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
		if key.ID == uuid.MustParse("00000000-0000-7000-0000-00000000ff77") {
			found = true
			if key.Name != "RoundTrip Key" || key.Provider != "openai" || key.SecretKey != "secret" {
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
	inserted, err := st.Models().InsertModel(ctx, types.ModelInfo{
		Provider:     types.ProviderCustom,
		Type:         "gpt-roundtrip",
		Name:         "GPT RoundTrip",
		InputPrice:   1,
		OutputPrice:  2,
		MaxContext:   128000,
		Enabled:      true,
		Capabilities: []string{"chat"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defaultModelID := inserted.ID
	fallbackModelID := contract.IDModel1
	rules := []types.RoutingRule{
		{
			ID:              contract.IDDept3,
			NodeID:          contract.IDDept3,
			NodeName:        "后端组",
			DefaultModelID:  &defaultModelID,
			FallbackModelID: &fallbackModelID,
			AllowedModelIDs: []uuid.UUID{inserted.ID, contract.IDModel1},
			Inherited:       false,
		},
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
		if model.Type == "gpt-roundtrip" {
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
			if len(rule.AllowedModelIDs) != 2 || rule.DefaultModelID == nil || *rule.DefaultModelID != defaultModelID {
				t.Fatalf("routing rule mismatch: %+v", rule)
			}
		}
	}
	if !foundRule {
		t.Fatal("routing rule not found after round-trip")
	}
}

func TestModelAllowlistReplaceRejectsUnknownModel(t *testing.T) {
	t.Parallel()
	st := testPostgresStore(t)
	ctx := testutil.Ctx()
	err := st.Models().Allowlist().Replace(ctx, types.AllowlistOwnerOrgNode, contract.IDDept3, []uuid.UUID{uuid.MustParse("00000000-0000-7000-0000-0000000f4240")})
	if err == nil {
		t.Fatal("expected error for unknown model")
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

func TestOrgNodeTreeMemberCounts(t *testing.T) {
	t.Parallel()
	st := testPostgresStore(t)
	ctx := testutil.Ctx()
	nodes, err := st.Org().Nodes().Tree(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) == 0 {
		t.Fatal("expected org tree")
	}
	if nodes[0].MemberCount <= 0 {
		t.Fatalf("expected root memberCount > 0, got %d", nodes[0].MemberCount)
	}
}
