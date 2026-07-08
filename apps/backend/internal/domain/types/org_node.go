package types

// OrgNode is the storage SSOT for the company org tree (table org_nodes).
// Physical columns named department_id (e.g. members.department_id) refer to OrgNode.ID;
// see docs/Backend-存储架构.md §7.
const (
	AllowlistOwnerPlatformKey = "platform_key"
	AllowlistOwnerOrgNode     = "org_node"
	AllowlistOwnerKeyApproval = "key_approval"
)

type OrgNode struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	ParentID         *string   `json:"parentId"`
	Children         []OrgNode `json:"children,omitempty"`
	MemberCount      int       `json:"memberCount"`
	ExternalID       *string   `json:"externalId,omitempty"`
	Source           *string   `json:"source,omitempty"`
	ManagerID        *string   `json:"managerId,omitempty"`
	Budget           float64   `json:"budget"`
	Consumed         float64   `json:"consumed"`
	ReservedPool     *float64  `json:"reservedPool,omitempty"`
	Period           string    `json:"period"`
	DefaultModelId   *int64    `json:"defaultModelId,omitempty"`
	FallbackModelId  *int64    `json:"fallbackModelId,omitempty"`
	RoutingInherited bool      `json:"routingInherited"`
}

func OrgNodeToDepartment(node OrgNode) Department {
	return Department{
		ID:          node.ID,
		Name:        node.Name,
		ParentID:    node.ParentID,
		Children:    orgNodeChildrenToDepartments(node.Children),
		MemberCount: node.MemberCount,
		ExternalID:  node.ExternalID,
		Source:      node.Source,
		ManagerID:   node.ManagerID,
	}
}

func OrgNodesToDepartments(nodes []OrgNode) []Department {
	result := make([]Department, len(nodes))
	for i, node := range nodes {
		result[i] = OrgNodeToDepartment(node)
	}
	return result
}

func orgNodeChildrenToDepartments(children []OrgNode) []Department {
	if len(children) == 0 {
		return nil
	}
	result := make([]Department, len(children))
	for i, child := range children {
		result[i] = OrgNodeToDepartment(child)
	}
	return result
}

func OrgNodeToBudgetNode(node OrgNode) BudgetNode {
	return BudgetNode{
		ID:           node.ID,
		Name:         node.Name,
		ParentID:     node.ParentID,
		Budget:       node.Budget,
		Consumed:     node.Consumed,
		ReservedPool: node.ReservedPool,
		Children:     orgNodeChildrenToBudgetNodes(node.Children),
		Period:       node.Period,
	}
}

func OrgNodesToBudgetTree(nodes []OrgNode) []BudgetNode {
	result := make([]BudgetNode, len(nodes))
	for i, node := range nodes {
		result[i] = OrgNodeToBudgetNode(node)
	}
	return result
}

func orgNodeChildrenToBudgetNodes(children []OrgNode) []BudgetNode {
	if len(children) == 0 {
		return nil
	}
	result := make([]BudgetNode, len(children))
	for i, child := range children {
		result[i] = OrgNodeToBudgetNode(child)
	}
	return result
}

func OrgNodeToRoutingRule(node OrgNode, allowedModelIDs []int64) RoutingRule {
	return RoutingRule{
		ID:              node.ID,
		NodeID:          node.ID,
		NodeName:        node.Name,
		AllowedModelIds: append([]int64{}, allowedModelIDs...),
		DefaultModelId:  node.DefaultModelId,
		FallbackModelId: node.FallbackModelId,
		Inherited:       node.RoutingInherited,
	}
}

func ApplyBudgetTreeToOrgNodes(nodes []OrgNode, tree []BudgetNode) {
	byID := make(map[string]BudgetNode)
	var walk func(items []BudgetNode)
	walk = func(items []BudgetNode) {
		for _, node := range items {
			flat := node
			flat.Children = nil
			byID[node.ID] = flat
			if len(node.Children) > 0 {
				walk(node.Children)
			}
		}
	}
	walk(tree)
	applyBudgetFields(nodes, byID)
}

func applyBudgetFields(nodes []OrgNode, byID map[string]BudgetNode) {
	for i := range nodes {
		if budget, ok := byID[nodes[i].ID]; ok {
			nodes[i].Budget = budget.Budget
			nodes[i].Consumed = budget.Consumed
			nodes[i].ReservedPool = budget.ReservedPool
			nodes[i].Period = budget.Period
		}
		if len(nodes[i].Children) > 0 {
			applyBudgetFields(nodes[i].Children, byID)
		}
	}
}
