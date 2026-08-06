package org

import (
	"context"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
)

// OrgNodeTreeReader is the narrow interface for reading org node trees.
// Implemented by store.OrgNodeRepository.
type OrgNodeTreeReader interface {
	Tree(ctx context.Context) ([]types.OrgNode, error)
}

// AllowlistReader is the narrow interface for reading model allowlists.
// Implemented by store.ModelAllowlistRepository.
type AllowlistReader interface {
	List(ctx context.Context, ownerType string, ownerID uuid.UUID) ([]uuid.UUID, error)
}

// AllowlistWriter extends AllowlistReader with write operations for PersistRoutingRules.
type AllowlistWriter interface {
	AllowlistReader
	Replace(ctx context.Context, ownerType string, ownerID uuid.UUID, modelIDs []uuid.UUID) error
	DeleteByOwner(ctx context.Context, ownerType string, ownerID uuid.UUID) error
}

// OrgNodeTreeWriter can write org node trees. Implemented by store.OrgNodeRepository.
type OrgNodeTreeWriter interface {
	OrgNodeTreeReader
	SetTree(ctx context.Context, tree []types.OrgNode) error
}

func LoadDepartments(ctx context.Context, orgNodes OrgNodeTreeReader) ([]types.Department, error) {
	nodes, err := orgNodes.Tree(ctx)
	if err != nil {
		return nil, err
	}
	return types.OrgNodesToDepartments(nodes), nil
}

func LoadBudgetTree(ctx context.Context, orgNodes OrgNodeTreeReader) ([]types.BudgetNode, error) {
	nodes, err := orgNodes.Tree(ctx)
	if err != nil {
		return nil, err
	}
	return types.OrgNodesToBudgetTree(nodes), nil
}

func LoadRoutingRules(ctx context.Context, orgNodes OrgNodeTreeReader, allowlist AllowlistReader) ([]types.RoutingRule, error) {
	nodes, err := orgNodes.Tree(ctx)
	if err != nil {
		return nil, err
	}
	rules := make([]types.RoutingRule, 0)
	for _, node := range FlattenOrgNodeTree(nodes) {
		allowed, err := allowlist.List(ctx, types.AllowlistOwnerOrgNode, node.ID)
		if err != nil {
			return nil, err
		}
		rules = append(rules, types.OrgNodeToRoutingRule(node, allowed))
	}
	return rules, nil
}

func HasOrgNodeRoutingConfig(node types.OrgNode, allowed []uuid.UUID) bool {
	return len(allowed) > 0 || node.DefaultModelID != nil || node.FallbackModelID != nil || node.RoutingInherited
}

func RoutingRulesFromNodes(nodes []types.OrgNode, allowlists map[uuid.UUID][]uuid.UUID) []types.RoutingRule {
	rules := make([]types.RoutingRule, 0)
	for _, node := range FlattenOrgNodeTree(nodes) {
		allowed := allowlists[node.ID]
		if !HasOrgNodeRoutingConfig(node, allowed) {
			continue
		}
		rules = append(rules, types.OrgNodeToRoutingRule(node, allowed))
	}
	return rules
}

// PersistRoutingRules writes routing rules back to the store. Caller passes
// the allowlist writer (e.g. store.Models().Allowlist()) and org node writer
// (e.g. store.Org().Nodes()) separately to avoid import cycles with the store package.
func PersistRoutingRules(ctx context.Context, allowlist AllowlistWriter, orgNodes OrgNodeTreeWriter, nodes []types.OrgNode, rules []types.RoutingRule) error {
	ruleByNode := make(map[uuid.UUID]types.RoutingRule, len(rules))
	for _, rule := range rules {
		ruleByNode[rule.NodeID] = rule
	}
	for nodeID, rule := range ruleByNode {
		node := FindOrgNode(nodes, nodeID)
		if node != nil {
			node.DefaultModelID = rule.DefaultModelID
			node.FallbackModelID = rule.FallbackModelID
			node.RoutingInherited = rule.Inherited
		}
		if err := allowlist.Replace(ctx, types.AllowlistOwnerOrgNode, nodeID, rule.AllowedModelIDs); err != nil {
			return err
		}
	}
	for _, node := range FlattenOrgNodeTree(nodes) {
		if _, ok := ruleByNode[node.ID]; ok {
			continue
		}
		if err := allowlist.DeleteByOwner(ctx, types.AllowlistOwnerOrgNode, node.ID); err != nil {
			return err
		}
	}
	return orgNodes.SetTree(ctx, nodes)
}
