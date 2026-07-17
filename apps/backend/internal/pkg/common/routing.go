package common

import (
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
)

func buildDeptParentMap(departments []types.Department) map[uuid.UUID]*uuid.UUID {
	result := make(map[uuid.UUID]*uuid.UUID)
	var walk func(nodes []types.Department)
	walk = func(nodes []types.Department) {
		for _, dept := range nodes {
			if dept.ParentID != nil {
				parentID := *dept.ParentID
				result[dept.ID] = &parentID
			} else {
				result[dept.ID] = nil
			}
			if len(dept.Children) > 0 {
				walk(dept.Children)
			}
		}
	}
	walk(departments)
	return result
}

func getRoutingRuleForDept(deptID uuid.UUID, rules []types.RoutingRule, parents map[uuid.UUID]*uuid.UUID) *types.RoutingRule {
	current := deptID
	for {
		for i := range rules {
			if rules[i].NodeID == current {
				return &rules[i]
			}
		}
		parent, ok := parents[current]
		if !ok || parent == nil {
			return nil
		}
		current = *parent
	}
}

func getParentDeptID(deptID uuid.UUID, parents map[uuid.UUID]*uuid.UUID) *uuid.UUID {
	return parents[deptID]
}

func ResolveDeptAllowedModelIDs(
	deptID uuid.UUID,
	departments []types.Department,
	rules []types.RoutingRule,
	models []types.ModelInfo,
) []uuid.UUID {
	parents := buildDeptParentMap(departments)
	rule := getRoutingRuleForDept(deptID, rules, parents)
	if rule == nil {
		return modelcatalog.EnabledModelIDs(models)
	}

	parentID := getParentDeptID(rule.NodeID, parents)
	var parentRule *types.RoutingRule
	if parentID != nil {
		for i := range rules {
			if rules[i].NodeID == *parentID {
				parentRule = &rules[i]
				break
			}
		}
	}

	allowedModelIDs := append([]uuid.UUID{}, rule.AllowedModelIDs...)
	if rule.Inherited && parentRule != nil {
		filtered := make([]uuid.UUID, 0)
		parentSet := make(map[uuid.UUID]struct{}, len(parentRule.AllowedModelIDs))
		for _, id := range parentRule.AllowedModelIDs {
			parentSet[id] = struct{}{}
		}
		for _, id := range allowedModelIDs {
			if _, ok := parentSet[id]; ok {
				filtered = append(filtered, id)
			}
		}
		allowedModelIDs = filtered
		if len(allowedModelIDs) == 0 {
			allowedModelIDs = append([]uuid.UUID{}, parentRule.AllowedModelIDs...)
		}
	}
	return modelcatalog.FilterEnabledIDs(models, allowedModelIDs)
}

func ValidateModelIDsForMember(
	memberID uuid.UUID,
	modelIDs []uuid.UUID,
	members []types.Member,
	departments []types.Department,
	rules []types.RoutingRule,
	modelCatalog []types.ModelInfo,
	notInDeptMessage string,
) *string {
	if len(modelIDs) == 0 {
		return nil
	}
	var member *types.Member
	for i := range members {
		if members[i].ID == memberID {
			member = &members[i]
			break
		}
	}
	if member == nil {
		return nil
	}
	allowed := ResolveDeptAllowedModelIDs(member.DepartmentID, departments, rules, modelCatalog)
	allowedSet := make(map[uuid.UUID]struct{}, len(allowed))
	for _, id := range allowed {
		allowedSet[id] = struct{}{}
	}
	for _, id := range modelIDs {
		if _, ok := allowedSet[id]; !ok {
			return &notInDeptMessage
		}
	}
	return nil
}

func GetRoutingRuleForDept(
	deptID uuid.UUID,
	rules []types.RoutingRule,
	departments []types.Department,
) *types.RoutingRule {
	parents := buildDeptParentMap(departments)
	return getRoutingRuleForDept(deptID, rules, parents)
}

func GetParentDeptID(deptID uuid.UUID, departments []types.Department) *uuid.UUID {
	parents := buildDeptParentMap(departments)
	return getParentDeptID(deptID, parents)
}

func ShrinkChildRoutingRules(
	parentNodeID uuid.UUID,
	parentAllowed []uuid.UUID,
	rules []types.RoutingRule,
	departments []types.Department,
) []types.RoutingRule {
	parents := buildDeptParentMap(departments)
	result := cloneRoutingRulesSlice(rules)
	shrinkChildRoutingRules(parentNodeID, parentAllowed, result, parents)
	return result
}

func cloneRoutingRulesSlice(rules []types.RoutingRule) []types.RoutingRule {
	result := make([]types.RoutingRule, len(rules))
	for i, rule := range rules {
		result[i] = types.RoutingRule{
			ID:              rule.ID,
			NodeID:          rule.NodeID,
			NodeName:        rule.NodeName,
			AllowedModelIDs: append([]uuid.UUID{}, rule.AllowedModelIDs...),
			Inherited:       rule.Inherited,
		}
		if rule.DefaultModelID != nil {
			defaultModelID := *rule.DefaultModelID
			result[i].DefaultModelID = &defaultModelID
		}
		if rule.FallbackModelID != nil {
			fallbackModelID := *rule.FallbackModelID
			result[i].FallbackModelID = &fallbackModelID
		}
	}
	return result
}

func AppendInheritedRule(
	rules []types.RoutingRule,
	deptID uuid.UUID, deptName string,
	parentAllowed []uuid.UUID,
	ruleID uuid.UUID,
) []types.RoutingRule {
	return append(rules, types.RoutingRule{
		ID:              ruleID,
		NodeID:          deptID,
		NodeName:        deptName,
		AllowedModelIDs: append([]uuid.UUID{}, parentAllowed...),
		Inherited:       true,
	})
}

func RemoveRuleByNodeID(rules []types.RoutingRule, nodeID uuid.UUID) []types.RoutingRule {
	filtered := make([]types.RoutingRule, 0, len(rules))
	for _, rule := range rules {
		if rule.NodeID != nodeID {
			filtered = append(filtered, rule)
		}
	}
	return filtered
}

func UpdateRuleNodeName(rules []types.RoutingRule, nodeID uuid.UUID, name string) []types.RoutingRule {
	for i := range rules {
		if rules[i].NodeID == nodeID {
			rules[i].NodeName = name
		}
	}
	return rules
}

func EnrichRoutingRule(rule types.RoutingRule, catalog []types.ModelInfo) types.RoutingRule {
	validIDs := modelcatalog.FilterValidIDs(catalog, rule.AllowedModelIDs)
	rule.AllowedModelIDs = validIDs
	rule.AllowedModels = modelcatalog.EnrichRefs(catalog, validIDs)
	if rule.DefaultModelID != nil {
		if ref := modelcatalog.EnrichRef(catalog, rule.DefaultModelID); ref != nil {
			rule.DefaultModel = ref
		} else {
			rule.DefaultModelID = nil
		}
	}
	if rule.FallbackModelID != nil {
		if ref := modelcatalog.EnrichRef(catalog, rule.FallbackModelID); ref != nil {
			rule.FallbackModel = ref
		} else {
			rule.FallbackModelID = nil
		}
	}
	return rule
}

func shrinkChildRoutingRules(
	parentNodeID uuid.UUID,
	parentAllowed []uuid.UUID,
	rules []types.RoutingRule,
	parents map[uuid.UUID]*uuid.UUID,
) {
	for i := range rules {
		parentID := getParentDeptID(rules[i].NodeID, parents)
		if parentID == nil || *parentID != parentNodeID {
			continue
		}
		filtered := make([]uuid.UUID, 0)
		parentSet := make(map[uuid.UUID]struct{}, len(parentAllowed))
		for _, id := range parentAllowed {
			parentSet[id] = struct{}{}
		}
		for _, id := range rules[i].AllowedModelIDs {
			if _, ok := parentSet[id]; ok {
				filtered = append(filtered, id)
			}
		}
		if len(filtered) == 0 && len(parentAllowed) > 0 {
			filtered = append([]uuid.UUID{}, parentAllowed...)
		}
		rules[i].AllowedModelIDs = filtered
		shrinkChildRoutingRules(rules[i].NodeID, rules[i].AllowedModelIDs, rules, parents)
	}
}
