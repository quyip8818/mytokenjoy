package common

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/modelcatalog"
)

func buildDeptParentMap(departments []types.Department) map[string]*string {
	result := make(map[string]*string)
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

func getRoutingRuleForDept(deptID string, rules []types.RoutingRule, parents map[string]*string) *types.RoutingRule {
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

func getParentDeptID(deptID string, parents map[string]*string) *string {
	return parents[deptID]
}

func ResolveDeptAllowedModelIDs(
	deptID string,
	departments []types.Department,
	rules []types.RoutingRule,
	models []types.ModelInfo,
) []int64 {
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

	allowedModelIDs := append([]int64{}, rule.AllowedModelIDs...)
	if rule.Inherited && parentRule != nil {
		filtered := make([]int64, 0)
		parentSet := make(map[int64]struct{}, len(parentRule.AllowedModelIDs))
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
			allowedModelIDs = append([]int64{}, parentRule.AllowedModelIDs...)
		}
	}
	return modelcatalog.FilterEnabledIDs(models, allowedModelIDs)
}

func ValidateModelIDsForMember(
	memberID string,
	modelIDs []int64,
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
	allowedSet := make(map[int64]struct{}, len(allowed))
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
	deptID string,
	rules []types.RoutingRule,
	departments []types.Department,
) *types.RoutingRule {
	parents := buildDeptParentMap(departments)
	return getRoutingRuleForDept(deptID, rules, parents)
}

func GetParentDeptID(deptID string, departments []types.Department) *string {
	parents := buildDeptParentMap(departments)
	return getParentDeptID(deptID, parents)
}

func ShrinkChildRoutingRules(
	parentNodeID string,
	parentAllowed []int64,
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
			AllowedModelIDs: append([]int64{}, rule.AllowedModelIDs...),
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
	deptID, deptName string,
	parentAllowed []int64,
	ruleID string,
) []types.RoutingRule {
	return append(rules, types.RoutingRule{
		ID:              ruleID,
		NodeID:          deptID,
		NodeName:        deptName,
		AllowedModelIDs: append([]int64{}, parentAllowed...),
		Inherited:       true,
	})
}

func RemoveRuleByNodeID(rules []types.RoutingRule, nodeID string) []types.RoutingRule {
	filtered := make([]types.RoutingRule, 0, len(rules))
	for _, rule := range rules {
		if rule.NodeID != nodeID {
			filtered = append(filtered, rule)
		}
	}
	return filtered
}

func UpdateRuleNodeName(rules []types.RoutingRule, nodeID, name string) []types.RoutingRule {
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
	parentNodeID string,
	parentAllowed []int64,
	rules []types.RoutingRule,
	parents map[string]*string,
) {
	for i := range rules {
		parentID := getParentDeptID(rules[i].NodeID, parents)
		if parentID == nil || *parentID != parentNodeID {
			continue
		}
		filtered := make([]int64, 0)
		parentSet := make(map[int64]struct{}, len(parentAllowed))
		for _, id := range parentAllowed {
			parentSet[id] = struct{}{}
		}
		for _, id := range rules[i].AllowedModelIDs {
			if _, ok := parentSet[id]; ok {
				filtered = append(filtered, id)
			}
		}
		if len(filtered) == 0 && len(parentAllowed) > 0 {
			filtered = append([]int64{}, parentAllowed...)
		}
		rules[i].AllowedModelIDs = filtered
		shrinkChildRoutingRules(rules[i].NodeID, rules[i].AllowedModelIDs, rules, parents)
	}
}
