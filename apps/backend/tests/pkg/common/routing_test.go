package common_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

func TestShrinkChildRoutingRules(t *testing.T) {
	parentID := "dept-1"
	childParentID := "dept-2"
	departments := []types.Department{
		{ID: "dept-1", Name: "Root"},
		{ID: "dept-2", Name: "Child", ParentID: &parentID},
		{ID: "dept-3", Name: "Grandchild", ParentID: &childParentID},
	}
	rules := []types.RoutingRule{
		{ID: "r-1", NodeID: "dept-2", AllowedModels: []string{"gpt-4o", "claude", "deepseek"}},
		{ID: "r-2", NodeID: "dept-3", AllowedModels: []string{"gpt-4o", "claude", "deepseek"}},
	}
	updated := common.ShrinkChildRoutingRules("dept-1", []string{"gpt-4o"}, rules, departments)
	if len(updated[0].AllowedModels) != 1 || updated[0].AllowedModels[0] != "gpt-4o" {
		t.Fatalf("expected child rule to shrink to gpt-4o, got %v", updated[0].AllowedModels)
	}
	if len(updated[1].AllowedModels) != 1 || updated[1].AllowedModels[0] != "gpt-4o" {
		t.Fatalf("expected grandchild rule to shrink to gpt-4o, got %v", updated[1].AllowedModels)
	}
}
