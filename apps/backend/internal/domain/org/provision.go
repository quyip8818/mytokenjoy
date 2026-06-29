package org

import (
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
)

const (
	RootDepartmentID     = "dept-1"
	DeptDeleteBlockedMsg = "请先移动或删除该部门下的子部门和成员"
)

type ProvisionInput struct {
	ID       string
	Name     string
	ParentID string
	Period   string
}

type ProvisionState struct {
	Departments []types.Department
	BudgetTree  []types.BudgetNode
	Rules       []types.RoutingRule
	Models      []types.ModelInfo
}

func ProvisionDepartment(state *ProvisionState, input ProvisionInput) error {
	if pkgorg.FindDepartment(state.Departments, input.ParentID) == nil {
		return fmt.Errorf("parent department not found")
	}
	if pkgorg.FindDepartment(state.Departments, input.ID) != nil {
		return fmt.Errorf("department already exists")
	}

	parentID := input.ParentID
	dept := types.Department{
		ID: input.ID, Name: input.Name, ParentID: &parentID, MemberCount: 0,
	}
	if !pkgorg.InsertDepartmentChild(state.Departments, input.ParentID, dept) {
		return fmt.Errorf("failed to insert department")
	}

	budgetNode := types.BudgetNode{
		ID: input.ID, Name: input.Name, ParentID: &parentID,
		Budget: 0, Consumed: 0, Period: input.Period,
	}
	if !budget.InsertBudgetNode(state.BudgetTree, input.ParentID, budgetNode) {
		return fmt.Errorf("failed to insert budget node")
	}

	parentAllowed := common.ResolveDeptAllowedModels(
		input.ParentID, state.Departments, state.Rules, state.Models,
	)
	state.Rules = common.AppendInheritedRule(
		state.Rules, input.ID, input.Name, parentAllowed, fmt.Sprintf("rr-%s", input.ID),
	)
	return nil
}

func DeprovisionDepartment(state *ProvisionState, deptID string) error {
	var removed bool
	state.Departments, removed = pkgorg.RemoveDepartment(state.Departments, deptID)
	if !removed {
		return fmt.Errorf("department not found")
	}

	var budgetRemoved bool
	state.BudgetTree, budgetRemoved = budget.RemoveBudgetNode(state.BudgetTree, deptID)
	if !budgetRemoved {
		return fmt.Errorf("budget node not found")
	}

	state.Rules = common.RemoveRuleByNodeID(state.Rules, deptID)
	return nil
}

func RenameDepartment(state *ProvisionState, deptID, name string) error {
	if !pkgorg.UpdateDepartmentName(state.Departments, deptID, name) {
		return fmt.Errorf("department not found")
	}
	if !budget.UpdateBudgetNodeName(state.BudgetTree, deptID, name) {
		return fmt.Errorf("budget node not found")
	}
	state.Rules = common.UpdateRuleNodeName(state.Rules, deptID, name)
	return nil
}

func RecalcDepartmentMemberCounts(
	departments []types.Department,
	members []types.Member,
) []types.Department {
	return pkgorg.RecalcMemberCounts(departments, members)
}
