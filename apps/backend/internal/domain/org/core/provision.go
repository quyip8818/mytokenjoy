package core

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
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
	Nodes          []types.OrgNode
	Models         []types.ModelInfo
	NodeAllowlists map[string][]int64
}

func LoadProvisionState(ctx context.Context, st store.Store, nodes []types.OrgNode) (*ProvisionState, error) {
	models, err := st.Models().Models(ctx)
	if err != nil {
		return nil, err
	}
	allowlists := make(map[string][]int64)
	for _, node := range pkgorg.FlattenOrgNodeTree(nodes) {
		allowed, err := st.Models().Allowlist().List(ctx, types.AllowlistOwnerOrgNode, node.ID)
		if err != nil {
			return nil, err
		}
		if common.HasOrgNodeRoutingConfig(node, allowed) {
			allowlists[node.ID] = allowed
		}
	}
	return &ProvisionState{
		Nodes:          nodes,
		Models:         models,
		NodeAllowlists: allowlists,
	}, nil
}

func rulesFromState(state *ProvisionState) []types.RoutingRule {
	return common.RoutingRulesFromNodes(state.Nodes, state.NodeAllowlists)
}

func DepartmentsFromState(state *ProvisionState) []types.Department {
	return types.OrgNodesToDepartments(state.Nodes)
}

func PersistProvisionState(ctx context.Context, st store.Store, state *ProvisionState) error {
	if err := st.Org().Nodes().SetTree(ctx, state.Nodes); err != nil {
		return err
	}
	flat := pkgorg.FlattenOrgNodeTree(state.Nodes)
	active := make(map[string]struct{}, len(state.NodeAllowlists))
	for _, node := range flat {
		if allowed, ok := state.NodeAllowlists[node.ID]; ok {
			active[node.ID] = struct{}{}
			if err := st.Models().Allowlist().Replace(ctx, types.AllowlistOwnerOrgNode, node.ID, allowed); err != nil {
				return err
			}
			continue
		}
		if err := st.Models().Allowlist().DeleteByOwner(ctx, types.AllowlistOwnerOrgNode, node.ID); err != nil {
			return err
		}
	}
	_ = active
	return nil
}

func ProvisionDepartment(state *ProvisionState, input ProvisionInput) error {
	if pkgorg.FindOrgNode(state.Nodes, input.ParentID) == nil {
		return fmt.Errorf("parent department not found")
	}
	if pkgorg.FindOrgNode(state.Nodes, input.ID) != nil {
		return fmt.Errorf("department already exists")
	}

	parentID := input.ParentID
	node := types.OrgNode{
		ID: input.ID, Name: input.Name, ParentID: &parentID, MemberCount: 0,
		Budget: 0, Consumed: 0, Period: input.Period, RoutingInherited: true,
	}
	if !pkgorg.InsertOrgNodeChild(state.Nodes, input.ParentID, node) {
		return fmt.Errorf("failed to insert department")
	}

	departments := DepartmentsFromState(state)
	rules := rulesFromState(state)
	parentAllowed := common.ResolveDeptAllowedModelIDs(
		input.ParentID, departments, rules, state.Models,
	)
	if state.NodeAllowlists == nil {
		state.NodeAllowlists = make(map[string][]int64)
	}
	state.NodeAllowlists[input.ID] = append([]int64{}, parentAllowed...)
	return nil
}

func DeprovisionDepartment(state *ProvisionState, deptID string) error {
	before := len(pkgorg.FlattenOrgNodeTree(state.Nodes))
	state.Nodes = pkgorg.RemoveOrgNodeByID(state.Nodes, deptID)
	if len(pkgorg.FlattenOrgNodeTree(state.Nodes)) == before {
		return fmt.Errorf("department not found")
	}
	delete(state.NodeAllowlists, deptID)
	return nil
}

func RenameDepartment(state *ProvisionState, deptID, name string) error {
	if pkgorg.FindOrgNode(state.Nodes, deptID) == nil {
		return fmt.Errorf("department not found")
	}
	pkgorg.UpdateOrgNodeName(state.Nodes, deptID, name)
	return nil
}

func RecalcDepartmentMemberCounts(nodes []types.OrgNode, members []types.Member) []types.OrgNode {
	departments := types.OrgNodesToDepartments(nodes)
	updated := pkgorg.RecalcMemberCounts(departments, members)
	return mergeMemberCountsIntoNodes(nodes, updated)
}

func mergeMemberCountsIntoNodes(nodes []types.OrgNode, departments []types.Department) []types.OrgNode {
	countByID := make(map[string]int)
	for _, dept := range pkgorg.FlattenDepartmentTree(departments) {
		countByID[dept.ID] = dept.MemberCount
	}
	applyMemberCounts(nodes, countByID)
	return nodes
}

func applyMemberCounts(nodes []types.OrgNode, countByID map[string]int) {
	for i := range nodes {
		if count, ok := countByID[nodes[i].ID]; ok {
			nodes[i].MemberCount = count
		}
		if len(nodes[i].Children) > 0 {
			applyMemberCounts(nodes[i].Children, countByID)
		}
	}
}
