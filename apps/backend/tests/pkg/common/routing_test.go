package common_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

var (
	dept1 = uuid.MustParse("00000000-0000-7000-0000-000000000d01")
	dept2 = uuid.MustParse("00000000-0000-7000-0000-000000000d02")
	dept3 = uuid.MustParse("00000000-0000-7000-0000-000000000d03")

	model1 = uuid.MustParse("00000000-0000-7000-0000-0000000000b1")
	model2 = uuid.MustParse("00000000-0000-7000-0000-0000000000b2")
	model3 = uuid.MustParse("00000000-0000-7000-0000-0000000000b3")
	model4 = uuid.MustParse("00000000-0000-7000-0000-0000000000b4")

	ruleA = uuid.MustParse("00000000-0000-7000-0000-00000000aa01")
	ruleB = uuid.MustParse("00000000-0000-7000-0000-00000000aa02")

	nodeA = uuid.MustParse("00000000-0000-7000-0000-00000000cc01")
	nodeB = uuid.MustParse("00000000-0000-7000-0000-00000000cc02")
	nodeC = uuid.MustParse("00000000-0000-7000-0000-00000000cc03")
)

func TestShrinkChildRoutingRules(t *testing.T) {
	t.Parallel()
	departments := []types.Department{
		{ID: dept1, Name: "Root"},
		{ID: dept2, Name: "Child", ParentID: &dept1},
		{ID: dept3, Name: "Grandchild", ParentID: &dept2},
	}
	rules := []types.RoutingRule{
		{ID: ruleA, NodeID: dept2, AllowedModelIDs: []uuid.UUID{model1, model2, model3}},
		{ID: ruleB, NodeID: dept3, AllowedModelIDs: []uuid.UUID{model1, model2, model3}},
	}
	updated := common.ShrinkChildRoutingRules(dept1, []uuid.UUID{model1}, rules, departments)
	if len(updated[0].AllowedModelIDs) != 1 || updated[0].AllowedModelIDs[0] != model1 {
		t.Fatalf("expected child rule to shrink to model1, got %v", updated[0].AllowedModelIDs)
	}
	if len(updated[1].AllowedModelIDs) != 1 || updated[1].AllowedModelIDs[0] != model1 {
		t.Fatalf("expected grandchild rule to shrink to model1, got %v", updated[1].AllowedModelIDs)
	}
}

func TestResolveDeptAllowedModelIDs_NoRules(t *testing.T) {
	t.Parallel()
	models := []types.ModelInfo{
		{ID: model1, Type: "gpt-4", Enabled: true},
		{ID: model2, Type: "disabled", Enabled: false},
		{ID: model3, Type: "claude", Enabled: true},
	}
	result := common.ResolveDeptAllowedModelIDs(dept1, nil, nil, models)
	if len(result) != 2 {
		t.Fatalf("expected 2 enabled models, got %d: %v", len(result), result)
	}
}

func TestResolveDeptAllowedModelIDs_WithRule(t *testing.T) {
	t.Parallel()
	departments := []types.Department{
		{ID: dept1, Name: "Root"},
	}
	rules := []types.RoutingRule{
		{NodeID: dept1, AllowedModelIDs: []uuid.UUID{model1, model3}},
	}
	models := []types.ModelInfo{
		{ID: model1, Type: "gpt-4", Enabled: true},
		{ID: model3, Type: "claude", Enabled: true},
		{ID: model4, Type: "other", Enabled: true},
	}
	result := common.ResolveDeptAllowedModelIDs(dept1, departments, rules, models)
	if len(result) != 2 {
		t.Fatalf("expected 2 models from rule, got %d: %v", len(result), result)
	}
}

func TestRemoveRuleByNodeID(t *testing.T) {
	t.Parallel()
	rules := []types.RoutingRule{
		{NodeID: nodeA},
		{NodeID: nodeB},
		{NodeID: nodeC},
	}
	result := common.RemoveRuleByNodeID(rules, nodeB)
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
	for _, r := range result {
		if r.NodeID == nodeB {
			t.Fatal("expected 'b' to be removed")
		}
	}
}

func TestUpdateRuleNodeName(t *testing.T) {
	t.Parallel()
	rules := []types.RoutingRule{
		{NodeID: nodeA, NodeName: "old-a"},
		{NodeID: nodeB, NodeName: "old-b"},
	}
	result := common.UpdateRuleNodeName(rules, nodeA, "new-a")
	if result[0].NodeName != "new-a" {
		t.Errorf("expected 'new-a', got %q", result[0].NodeName)
	}
	if result[1].NodeName != "old-b" {
		t.Errorf("expected 'old-b' unchanged, got %q", result[1].NodeName)
	}
}

func TestGetRoutingRuleForDept_Inherited(t *testing.T) {
	t.Parallel()
	departments := []types.Department{
		{ID: dept1, Name: "Parent"},
		{ID: dept2, Name: "Child", ParentID: &dept1},
	}
	rules := []types.RoutingRule{
		{NodeID: dept1, AllowedModelIDs: []uuid.UUID{model1}},
	}
	rule := common.GetRoutingRuleForDept(dept2, rules, departments)
	if rule == nil {
		t.Fatal("expected to find parent rule")
	}
	if rule.NodeID != dept1 {
		t.Errorf("expected parent rule, got %q", rule.NodeID)
	}
}

func TestGetParentDeptID(t *testing.T) {
	t.Parallel()
	departments := []types.Department{
		{ID: dept1, Name: "Parent"},
		{ID: dept2, Name: "Child", ParentID: &dept1},
	}

	parent := common.GetParentDeptID(dept2, departments)
	if parent == nil || *parent != dept1 {
		t.Errorf("expected parent dept1, got %v", parent)
	}

	root := common.GetParentDeptID(dept1, departments)
	if root != nil {
		t.Errorf("expected nil for root, got %v", *root)
	}
}
