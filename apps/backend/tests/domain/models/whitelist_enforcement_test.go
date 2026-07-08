package models_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

func TestValidateModelIDsForMember_AllowedModel(t *testing.T) {
	t.Parallel()
	members := []types.Member{{ID: "m1", DepartmentID: "d1"}}
	departments := []types.Department{{ID: "d1", Name: "Dept"}}
	rules := []types.RoutingRule{{NodeID: "d1", AllowedModelIds: []int64{1, 2}}}
	models := []types.ModelInfo{
		{ModelID: 1, Type: "gpt-4", Enabled: true},
		{ModelID: 2, Type: "claude", Enabled: true},
	}
	errMsg := common.ValidateModelIDsForMember("m1", []int64{1}, members, departments, rules, models, "模型不可用")
	if errMsg != nil {
		t.Errorf("expected nil error for allowed model, got %q", *errMsg)
	}
}

func TestValidateModelIDsForMember_BlockedModel(t *testing.T) {
	t.Parallel()
	members := []types.Member{{ID: "m1", DepartmentID: "d1"}}
	departments := []types.Department{{ID: "d1", Name: "Dept"}}
	rules := []types.RoutingRule{{NodeID: "d1", AllowedModelIds: []int64{1}}}
	models := []types.ModelInfo{
		{ModelID: 1, Type: "gpt-4", Enabled: true},
		{ModelID: 2, Type: "claude", Enabled: true},
	}
	errMsg := common.ValidateModelIDsForMember("m1", []int64{2}, members, departments, rules, models, "模型不可用")
	if errMsg == nil {
		t.Fatal("expected error for blocked model")
	}
	if *errMsg != "模型不可用" {
		t.Errorf("unexpected error: %q", *errMsg)
	}
}

func TestValidateModelIDsForMember_EmptyWhitelist(t *testing.T) {
	t.Parallel()
	members := []types.Member{{ID: "m1", DepartmentID: "d1"}}
	models := []types.ModelInfo{
		{ModelID: 1, Type: "gpt-4", Enabled: true},
		{ModelID: 2, Type: "claude", Enabled: true},
	}
	errMsg := common.ValidateModelIDsForMember("m1", []int64{1}, members, nil, nil, models, "模型不可用")
	if errMsg != nil {
		t.Errorf("no rules should allow all enabled models, got %q", *errMsg)
	}
}

func TestValidateModelIDsForMember_NoModelsProvided(t *testing.T) {
	t.Parallel()
	members := []types.Member{{ID: "m1", DepartmentID: "d1"}}
	if errMsg := common.ValidateModelIDsForMember("m1", nil, members, nil, nil, nil, "err"); errMsg != nil {
		t.Error("nil models should pass validation")
	}
	if errMsg := common.ValidateModelIDsForMember("m1", []int64{}, members, nil, nil, nil, "err"); errMsg != nil {
		t.Error("empty models should pass validation")
	}
}

func TestWhitelistInheritance_ChildNarrowsFromParent(t *testing.T) {
	t.Parallel()
	parentID := "d-parent"
	departments := []types.Department{
		{ID: "d-parent", Name: "Parent"},
		{ID: "d-child", Name: "Child", ParentID: &parentID},
	}
	rules := []types.RoutingRule{
		{NodeID: "d-parent", AllowedModelIds: []int64{1, 2, 3}},
		{NodeID: "d-child", AllowedModelIds: []int64{1, 2}, Inherited: true},
	}
	models := []types.ModelInfo{
		{ModelID: 1, Type: "gpt-4", Enabled: true},
		{ModelID: 2, Type: "claude", Enabled: true},
		{ModelID: 3, Type: "deepseek", Enabled: true},
	}
	member := []types.Member{{ID: "m-child", DepartmentID: "d-child"}}
	if errMsg := common.ValidateModelIDsForMember("m-child", []int64{1}, member, departments, rules, models, "模型不可用"); errMsg != nil {
		t.Errorf("child should access gpt-4: %q", *errMsg)
	}
	if errMsg := common.ValidateModelIDsForMember("m-child", []int64{3}, member, departments, rules, models, "模型不可用"); errMsg == nil {
		t.Error("child should NOT access deepseek (narrowed from parent)")
	}
}

func TestWhitelistInheritance_ParentShrinkSyncsChild(t *testing.T) {
	t.Parallel()
	parentID := "d-parent"
	departments := []types.Department{
		{ID: "d-parent", Name: "Parent"},
		{ID: "d-child", Name: "Child", ParentID: &parentID},
	}
	rules := []types.RoutingRule{
		{NodeID: "d-child", AllowedModelIds: []int64{1, 2, 3}, Inherited: true},
	}
	updated := common.ShrinkChildRoutingRules("d-parent", []int64{1}, rules, departments)
	if len(updated[0].AllowedModelIds) != 1 {
		t.Fatalf("expected child shrunk to 1 model, got %d: %v", len(updated[0].AllowedModelIds), updated[0].AllowedModelIds)
	}
	if updated[0].AllowedModelIds[0] != 1 {
		t.Errorf("expected model 1, got %d", updated[0].AllowedModelIds[0])
	}
}

func TestResolveDeptAllowedModelIDs_InheritedFromParent(t *testing.T) {
	t.Parallel()
	parentID := "d-parent"
	departments := []types.Department{
		{ID: "d-parent", Name: "Parent"},
		{ID: "d-child", Name: "Child", ParentID: &parentID},
	}
	rules := []types.RoutingRule{
		{NodeID: "d-parent", AllowedModelIds: []int64{1, 2}},
	}
	models := []types.ModelInfo{
		{ModelID: 1, Type: "gpt-4", Enabled: true},
		{ModelID: 2, Type: "claude", Enabled: true},
		{ModelID: 3, Type: "other", Enabled: true},
	}
	allowed := common.ResolveDeptAllowedModelIDs("d-child", departments, rules, models)
	if len(allowed) != 2 {
		t.Fatalf("expected 2 inherited models, got %d: %v", len(allowed), allowed)
	}
}
