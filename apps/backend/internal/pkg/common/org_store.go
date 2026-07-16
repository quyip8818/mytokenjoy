package common

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

func LoadDepartments(ctx context.Context, orgNodes store.OrgNodeRepository) ([]types.Department, error) {
	nodes, err := orgNodes.Tree(ctx)
	if err != nil {
		return nil, err
	}
	return types.OrgNodesToDepartments(nodes), nil
}

func LoadBudgetTree(ctx context.Context, orgNodes store.OrgNodeRepository) ([]types.BudgetNode, error) {
	nodes, err := orgNodes.Tree(ctx)
	if err != nil {
		return nil, err
	}
	return types.OrgNodesToBudgetTree(nodes), nil
}

func LoadRoutingRules(ctx context.Context, orgNodes store.OrgNodeRepository, allowlist store.ModelAllowlistRepository) ([]types.RoutingRule, error) {
	nodes, err := orgNodes.Tree(ctx)
	if err != nil {
		return nil, err
	}
	rules := make([]types.RoutingRule, 0)
	for _, node := range pkgorg.FlattenOrgNodeTree(nodes) {
		allowed, err := allowlist.List(ctx, types.AllowlistOwnerOrgNode, node.ID)
		if err != nil {
			return nil, err
		}
		if !HasOrgNodeRoutingConfig(node, allowed) {
			continue
		}
		rules = append(rules, types.OrgNodeToRoutingRule(node, allowed))
	}
	return rules, nil
}

func HasOrgNodeRoutingConfig(node types.OrgNode, allowed []int64) bool {
	return len(allowed) > 0 || node.DefaultModelID != nil || node.FallbackModelID != nil || node.RoutingInherited
}

func RoutingRulesFromNodes(nodes []types.OrgNode, allowlists map[string][]int64) []types.RoutingRule {
	rules := make([]types.RoutingRule, 0)
	for _, node := range pkgorg.FlattenOrgNodeTree(nodes) {
		allowed := allowlists[node.ID]
		if !HasOrgNodeRoutingConfig(node, allowed) {
			continue
		}
		rules = append(rules, types.OrgNodeToRoutingRule(node, allowed))
	}
	return rules
}

// RoutingPersistStore is the minimal store surface PersistRoutingRules needs.
type RoutingPersistStore interface {
	Models() store.ModelsRepository
	Org() store.OrgRepository
}

func PersistRoutingRules(ctx context.Context, st RoutingPersistStore, nodes []types.OrgNode, rules []types.RoutingRule) error {
	ruleByNode := make(map[string]types.RoutingRule, len(rules))
	for _, rule := range rules {
		ruleByNode[rule.NodeID] = rule
	}
	for nodeID, rule := range ruleByNode {
		node := pkgorg.FindOrgNode(nodes, nodeID)
		if node != nil {
			node.DefaultModelID = rule.DefaultModelID
			node.FallbackModelID = rule.FallbackModelID
			node.RoutingInherited = rule.Inherited
		}
		if err := st.Models().Allowlist().Replace(ctx, types.AllowlistOwnerOrgNode, nodeID, rule.AllowedModelIDs); err != nil {
			return err
		}
	}
	for _, node := range pkgorg.FlattenOrgNodeTree(nodes) {
		if _, ok := ruleByNode[node.ID]; ok {
			continue
		}
		if err := st.Models().Allowlist().DeleteByOwner(ctx, types.AllowlistOwnerOrgNode, node.ID); err != nil {
			return err
		}
	}
	return st.Org().Nodes().SetTree(ctx, nodes)
}
