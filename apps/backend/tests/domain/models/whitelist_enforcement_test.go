package models_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

func TestValidateModelIDsForMember_AllowedModel(t *testing.T) {
	t.Parallel()
	m1 := uuid.MustParse("00000000-0000-7000-0000-000000000001")
	d1 := uuid.MustParse("00000000-0000-7000-0000-000000000011")
	model1 := uuid.MustParse("00000000-0000-7000-0000-000000000101")
	model2 := uuid.MustParse("00000000-0000-7000-0000-000000000102")
	members := []types.Member{{ID: m1, DepartmentID: d1}}
	departments := []types.Department{{ID: d1, Name: "Dept"}}
	rules := []types.RoutingRule{{NodeID: d1, AllowedModelIDs: []uuid.UUID{model1, model2}}}
	models := []types.ModelInfo{
		{ID: model1, Type: "gpt-4", Enabled: true},
		{ID: model2, Type: "claude", Enabled: true},
	}
	errMsg := common.ValidateModelIDsForMember(m1, []uuid.UUID{model1}, members, departments, rules, models, "模型不可用")
	if errMsg != nil {
		t.Errorf("expected nil error for allowed model, got %q", *errMsg)
	}
}

func TestValidateModelIDsForMember_BlockedModel(t *testing.T) {
	t.Parallel()
	m1 := uuid.MustParse("00000000-0000-7000-0000-000000000001")
	d1 := uuid.MustParse("00000000-0000-7000-0000-000000000011")
	model1 := uuid.MustParse("00000000-0000-7000-0000-000000000101")
	model2 := uuid.MustParse("00000000-0000-7000-0000-000000000102")
	members := []types.Member{{ID: m1, DepartmentID: d1}}
	departments := []types.Department{{ID: d1, Name: "Dept"}}
	rules := []types.RoutingRule{{NodeID: d1, AllowedModelIDs: []uuid.UUID{model1}}}
	models := []types.ModelInfo{
		{ID: model1, Type: "gpt-4", Enabled: true},
		{ID: model2, Type: "claude", Enabled: true},
	}
	errMsg := common.ValidateModelIDsForMember(m1, []uuid.UUID{model2}, members, departments, rules, models, "模型不可用")
	if errMsg == nil {
		t.Fatal("expected error for blocked model")
	}
	if *errMsg != "模型不可用" {
		t.Errorf("unexpected error: %q", *errMsg)
	}
}

func TestValidateModelIDsForMember_EmptyWhitelist(t *testing.T) {
	t.Parallel()
	m1 := uuid.MustParse("00000000-0000-7000-0000-000000000001")
	d1 := uuid.MustParse("00000000-0000-7000-0000-000000000011")
	model1 := uuid.MustParse("00000000-0000-7000-0000-000000000101")
	model2 := uuid.MustParse("00000000-0000-7000-0000-000000000102")
	members := []types.Member{{ID: m1, DepartmentID: d1}}
	models := []types.ModelInfo{
		{ID: model1, Type: "gpt-4", Enabled: true},
		{ID: model2, Type: "claude", Enabled: true},
	}
	errMsg := common.ValidateModelIDsForMember(m1, []uuid.UUID{model1}, members, nil, nil, models, "模型不可用")
	if errMsg != nil {
		t.Errorf("no rules should allow all enabled models, got %q", *errMsg)
	}
}

func TestValidateModelIDsForMember_NoModelsProvided(t *testing.T) {
	t.Parallel()
	m1 := uuid.MustParse("00000000-0000-7000-0000-000000000001")
	d1 := uuid.MustParse("00000000-0000-7000-0000-000000000011")
	members := []types.Member{{ID: m1, DepartmentID: d1}}
	if errMsg := common.ValidateModelIDsForMember(m1, nil, members, nil, nil, nil, "err"); errMsg != nil {
		t.Error("nil models should pass validation")
	}
	if errMsg := common.ValidateModelIDsForMember(m1, []uuid.UUID{}, members, nil, nil, nil, "err"); errMsg != nil {
		t.Error("empty models should pass validation")
	}
}

func TestWhitelistInheritance_ChildNarrowsFromParent(t *testing.T) {
	t.Parallel()
	dParent := uuid.MustParse("00000000-0000-7000-0000-000000000021")
	dChild := uuid.MustParse("00000000-0000-7000-0000-000000000022")
	mChild := uuid.MustParse("00000000-0000-7000-0000-000000000031")
	model1 := uuid.MustParse("00000000-0000-7000-0000-000000000101")
	model2 := uuid.MustParse("00000000-0000-7000-0000-000000000102")
	model3 := uuid.MustParse("00000000-0000-7000-0000-000000000103")
	departments := []types.Department{
		{ID: dParent, Name: "Parent"},
		{ID: dChild, Name: "Child", ParentID: &dParent},
	}
	rules := []types.RoutingRule{
		{NodeID: dParent, AllowedModelIDs: []uuid.UUID{model1, model2, model3}},
		{NodeID: dChild, AllowedModelIDs: []uuid.UUID{model1, model2}, Inherited: true},
	}
	models := []types.ModelInfo{
		{ID: model1, Type: "gpt-4", Enabled: true},
		{ID: model2, Type: "claude", Enabled: true},
		{ID: model3, Type: "deepseek", Enabled: true},
	}
	member := []types.Member{{ID: mChild, DepartmentID: dChild}}
	if errMsg := common.ValidateModelIDsForMember(mChild, []uuid.UUID{model1}, member, departments, rules, models, "模型不可用"); errMsg != nil {
		t.Errorf("child should access gpt-4: %q", *errMsg)
	}
	if errMsg := common.ValidateModelIDsForMember(mChild, []uuid.UUID{model3}, member, departments, rules, models, "模型不可用"); errMsg == nil {
		t.Error("child should NOT access deepseek (narrowed from parent)")
	}
}

func TestWhitelistInheritance_ParentShrinkSyncsChild(t *testing.T) {
	t.Parallel()
	dParent := uuid.MustParse("00000000-0000-7000-0000-000000000021")
	dChild := uuid.MustParse("00000000-0000-7000-0000-000000000022")
	model1 := uuid.MustParse("00000000-0000-7000-0000-000000000101")
	model2 := uuid.MustParse("00000000-0000-7000-0000-000000000102")
	model3 := uuid.MustParse("00000000-0000-7000-0000-000000000103")
	departments := []types.Department{
		{ID: dParent, Name: "Parent"},
		{ID: dChild, Name: "Child", ParentID: &dParent},
	}
	rules := []types.RoutingRule{
		{NodeID: dChild, AllowedModelIDs: []uuid.UUID{model1, model2, model3}, Inherited: true},
	}
	updated := common.ShrinkChildRoutingRules(dParent, []uuid.UUID{model1}, rules, departments)
	if len(updated[0].AllowedModelIDs) != 1 {
		t.Fatalf("expected child shrunk to 1 model, got %d: %v", len(updated[0].AllowedModelIDs), updated[0].AllowedModelIDs)
	}
	if updated[0].AllowedModelIDs[0] != model1 {
		t.Errorf("expected model1, got %s", updated[0].AllowedModelIDs[0])
	}
}

func TestResolveDeptAllowedModelIDs_InheritedFromParent(t *testing.T) {
	t.Parallel()
	dParent := uuid.MustParse("00000000-0000-7000-0000-000000000021")
	dChild := uuid.MustParse("00000000-0000-7000-0000-000000000022")
	model1 := uuid.MustParse("00000000-0000-7000-0000-000000000101")
	model2 := uuid.MustParse("00000000-0000-7000-0000-000000000102")
	model3 := uuid.MustParse("00000000-0000-7000-0000-000000000103")
	departments := []types.Department{
		{ID: dParent, Name: "Parent"},
		{ID: dChild, Name: "Child", ParentID: &dParent},
	}
	rules := []types.RoutingRule{
		{NodeID: dParent, AllowedModelIDs: []uuid.UUID{model1, model2}},
	}
	models := []types.ModelInfo{
		{ID: model1, Type: "gpt-4", Enabled: true},
		{ID: model2, Type: "claude", Enabled: true},
		{ID: model3, Type: "other", Enabled: true},
	}
	allowed := common.ResolveDeptAllowedModelIDs(dChild, departments, rules, models)
	if len(allowed) != 2 {
		t.Fatalf("expected 2 inherited models, got %d: %v", len(allowed), allowed)
	}
}
