package common_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

func TestShrinkChildRoutingRules(t *testing.T) {
	t.Parallel()
	parentID := "dept-1"
	childParentID := "dept-2"
	departments := []types.Department{
		{ID: "dept-1", Name: "Root"},
		{ID: "dept-2", Name: "Child", ParentID: &parentID},
		{ID: "dept-3", Name: "Grandchild", ParentID: &childParentID},
	}
	rules := []types.RoutingRule{
		{ID: "r-1", NodeID: "dept-2", AllowedModelIDs: []int64{1, 2, 3}},
		{ID: "r-2", NodeID: "dept-3", AllowedModelIDs: []int64{1, 2, 3}},
	}
	updated := common.ShrinkChildRoutingRules("dept-1", []int64{1}, rules, departments)
	if len(updated[0].AllowedModelIDs) != 1 || updated[0].AllowedModelIDs[0] != 1 {
		t.Fatalf("expected child rule to shrink to model 1, got %v", updated[0].AllowedModelIDs)
	}
	if len(updated[1].AllowedModelIDs) != 1 || updated[1].AllowedModelIDs[0] != 1 {
		t.Fatalf("expected grandchild rule to shrink to model 1, got %v", updated[1].AllowedModelIDs)
	}
}

func TestResolveDeptAllowedModelIDs_NoRules(t *testing.T) {
	t.Parallel()
	models := []types.ModelInfo{
		{ModelID: 1, Type: "gpt-4", Enabled: true},
		{ModelID: 2, Type: "disabled", Enabled: false},
		{ModelID: 3, Type: "claude", Enabled: true},
	}
	result := common.ResolveDeptAllowedModelIDs("dept-1", nil, nil, models)
	if len(result) != 2 {
		t.Fatalf("expected 2 enabled models, got %d: %v", len(result), result)
	}
}

func TestResolveDeptAllowedModelIDs_WithRule(t *testing.T) {
	t.Parallel()
	departments := []types.Department{
		{ID: "dept-1", Name: "Root"},
	}
	rules := []types.RoutingRule{
		{NodeID: "dept-1", AllowedModelIDs: []int64{1, 3}},
	}
	models := []types.ModelInfo{
		{ModelID: 1, Type: "gpt-4", Enabled: true},
		{ModelID: 3, Type: "claude", Enabled: true},
		{ModelID: 4, Type: "other", Enabled: true},
	}
	result := common.ResolveDeptAllowedModelIDs("dept-1", departments, rules, models)
	if len(result) != 2 {
		t.Fatalf("expected 2 models from rule, got %d: %v", len(result), result)
	}
}

func TestRemoveRuleByNodeID(t *testing.T) {
	t.Parallel()
	rules := []types.RoutingRule{
		{NodeID: "a"},
		{NodeID: "b"},
		{NodeID: "c"},
	}
	result := common.RemoveRuleByNodeID(rules, "b")
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
	for _, r := range result {
		if r.NodeID == "b" {
			t.Fatal("expected 'b' to be removed")
		}
	}
}

func TestUpdateRuleNodeName(t *testing.T) {
	t.Parallel()
	rules := []types.RoutingRule{
		{NodeID: "a", NodeName: "old-a"},
		{NodeID: "b", NodeName: "old-b"},
	}
	result := common.UpdateRuleNodeName(rules, "a", "new-a")
	if result[0].NodeName != "new-a" {
		t.Errorf("expected 'new-a', got %q", result[0].NodeName)
	}
	if result[1].NodeName != "old-b" {
		t.Errorf("expected 'old-b' unchanged, got %q", result[1].NodeName)
	}
}

func TestGetRoutingRuleForDept_Inherited(t *testing.T) {
	t.Parallel()
	parentID := "dept-1"
	departments := []types.Department{
		{ID: "dept-1", Name: "Parent"},
		{ID: "dept-2", Name: "Child", ParentID: &parentID},
	}
	rules := []types.RoutingRule{
		{NodeID: "dept-1", AllowedModelIDs: []int64{1}},
	}
	rule := common.GetRoutingRuleForDept("dept-2", rules, departments)
	if rule == nil {
		t.Fatal("expected to find parent rule")
	}
	if rule.NodeID != "dept-1" {
		t.Errorf("expected parent rule, got %q", rule.NodeID)
	}
}

func TestGetParentDeptID(t *testing.T) {
	t.Parallel()
	parentID := "dept-1"
	departments := []types.Department{
		{ID: "dept-1", Name: "Parent"},
		{ID: "dept-2", Name: "Child", ParentID: &parentID},
	}

	parent := common.GetParentDeptID("dept-2", departments)
	if parent == nil || *parent != "dept-1" {
		t.Errorf("expected parent 'dept-1', got %v", parent)
	}

	root := common.GetParentDeptID("dept-1", departments)
	if root != nil {
		t.Errorf("expected nil for root, got %v", *root)
	}
}
