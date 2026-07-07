package models_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

// PRD 4.2: Model whitelist enforcement rules
// - "API 调用时指定模型不在白名单内 → 返回错误"
// - "只能缩小不能扩大"
// - "父级缩小 → 子级自动同步缩小"

func TestValidateModelsForMember_AllowedModel(t *testing.T) {
	members := []types.Member{
		{ID: "m1", DepartmentID: "d1"},
	}
	departments := []types.Department{
		{ID: "d1", Name: "Dept"},
	}
	rules := []types.RoutingRule{
		{NodeID: "d1", AllowedModels: []string{"gpt-4", "claude"}},
	}
	models := []types.ModelInfo{
		{Name: "gpt-4", Enabled: true},
		{Name: "claude", Enabled: true},
	}

	errMsg := common.ValidateModelsForMember("m1", []string{"gpt-4"}, members, departments, rules, models, "模型不可用")
	if errMsg != nil {
		t.Errorf("expected nil error for allowed model, got %q", *errMsg)
	}
}

func TestValidateModelsForMember_BlockedModel(t *testing.T) {
	members := []types.Member{
		{ID: "m1", DepartmentID: "d1"},
	}
	departments := []types.Department{
		{ID: "d1", Name: "Dept"},
	}
	rules := []types.RoutingRule{
		{NodeID: "d1", AllowedModels: []string{"gpt-4"}},
	}
	models := []types.ModelInfo{
		{Name: "gpt-4", Enabled: true},
		{Name: "claude", Enabled: true},
	}

	errMsg := common.ValidateModelsForMember("m1", []string{"claude"}, members, departments, rules, models, "模型不可用")
	if errMsg == nil {
		t.Fatal("expected error for blocked model")
	}
	if *errMsg != "模型不可用" {
		t.Errorf("unexpected error: %q", *errMsg)
	}
}

func TestValidateModelsForMember_EmptyWhitelist(t *testing.T) {
	members := []types.Member{
		{ID: "m1", DepartmentID: "d1"},
	}
	// No routing rule means all enabled models are allowed
	models := []types.ModelInfo{
		{Name: "gpt-4", Enabled: true},
		{Name: "claude", Enabled: true},
	}

	errMsg := common.ValidateModelsForMember("m1", []string{"gpt-4"}, members, nil, nil, models, "模型不可用")
	if errMsg != nil {
		t.Errorf("no rules should allow all enabled models, got %q", *errMsg)
	}
}

func TestValidateModelsForMember_NoModelsProvided(t *testing.T) {
	members := []types.Member{{ID: "m1", DepartmentID: "d1"}}

	// Empty model list should always be valid
	errMsg := common.ValidateModelsForMember("m1", nil, members, nil, nil, nil, "err")
	if errMsg != nil {
		t.Error("nil models should pass validation")
	}

	errMsg = common.ValidateModelsForMember("m1", []string{}, members, nil, nil, nil, "err")
	if errMsg != nil {
		t.Error("empty models should pass validation")
	}
}

func TestWhitelistInheritance_ChildNarrowsFromParent(t *testing.T) {
	parentID := "d-parent"
	departments := []types.Department{
		{ID: "d-parent", Name: "Parent"},
		{ID: "d-child", Name: "Child", ParentID: &parentID},
	}
	rules := []types.RoutingRule{
		{NodeID: "d-parent", AllowedModels: []string{"gpt-4", "claude", "deepseek"}},
		{NodeID: "d-child", AllowedModels: []string{"gpt-4", "claude"}, Inherited: true},
	}
	models := []types.ModelInfo{
		{Name: "gpt-4", Enabled: true},
		{Name: "claude", Enabled: true},
		{Name: "deepseek", Enabled: true},
	}

	// Child member can use gpt-4 (in both parent and child whitelist)
	errMsg := common.ValidateModelsForMember("m-child", []string{"gpt-4"},
		[]types.Member{{ID: "m-child", DepartmentID: "d-child"}},
		departments, rules, models, "模型不可用")
	if errMsg != nil {
		t.Errorf("child should access gpt-4: %q", *errMsg)
	}

	// Child member CANNOT use deepseek (in parent but not in child)
	errMsg = common.ValidateModelsForMember("m-child", []string{"deepseek"},
		[]types.Member{{ID: "m-child", DepartmentID: "d-child"}},
		departments, rules, models, "模型不可用")
	if errMsg == nil {
		t.Error("child should NOT access deepseek (narrowed from parent)")
	}
}

func TestWhitelistInheritance_ParentShrinkSyncsChild(t *testing.T) {
	parentID := "d-parent"
	departments := []types.Department{
		{ID: "d-parent", Name: "Parent"},
		{ID: "d-child", Name: "Child", ParentID: &parentID},
	}
	rules := []types.RoutingRule{
		{NodeID: "d-child", AllowedModels: []string{"gpt-4", "claude", "deepseek"}, Inherited: true},
	}

	// Parent shrinks to only gpt-4 → child should also shrink
	updated := common.ShrinkChildRoutingRules("d-parent", []string{"gpt-4"}, rules, departments)
	if len(updated[0].AllowedModels) != 1 {
		t.Fatalf("expected child shrunk to 1 model, got %d: %v", len(updated[0].AllowedModels), updated[0].AllowedModels)
	}
	if updated[0].AllowedModels[0] != "gpt-4" {
		t.Errorf("expected gpt-4, got %s", updated[0].AllowedModels[0])
	}
}

func TestResolveDeptAllowedModels_InheritedFromParent(t *testing.T) {
	parentID := "d-parent"
	departments := []types.Department{
		{ID: "d-parent", Name: "Parent"},
		{ID: "d-child", Name: "Child", ParentID: &parentID},
	}
	rules := []types.RoutingRule{
		{NodeID: "d-parent", AllowedModels: []string{"gpt-4", "claude"}},
	}
	models := []types.ModelInfo{
		{Name: "gpt-4", Enabled: true},
		{Name: "claude", Enabled: true},
		{Name: "other", Enabled: true},
	}

	// Child with no rule inherits from parent
	allowed := common.ResolveDeptAllowedModels("d-child", departments, rules, models)
	if len(allowed) != 2 {
		t.Fatalf("expected 2 inherited models, got %d: %v", len(allowed), allowed)
	}
}
