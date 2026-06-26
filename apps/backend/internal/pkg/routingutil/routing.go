package routingutil

import (
	"github.com/tokenjoy/backend/internal/domain/types"
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

func ResolveDeptAllowedModels(
	deptID string,
	departments []types.Department,
	rules []types.RoutingRule,
	models []types.ModelInfo,
) []string {
	parents := buildDeptParentMap(departments)
	rule := getRoutingRuleForDept(deptID, rules, parents)
	if rule == nil {
		allowed := make([]string, 0)
		for _, model := range models {
			if model.Enabled {
				allowed = append(allowed, model.Name)
			}
		}
		return allowed
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

	allowedModels := append([]string{}, rule.AllowedModels...)
	if rule.Inherited && parentRule != nil {
		filtered := make([]string, 0)
		for _, model := range allowedModels {
			for _, parentModel := range parentRule.AllowedModels {
				if model == parentModel {
					filtered = append(filtered, model)
					break
				}
			}
		}
		allowedModels = filtered
		if len(allowedModels) == 0 {
			allowedModels = append([]string{}, parentRule.AllowedModels...)
		}
	}
	return allowedModels
}

func ValidateModelsForMember(
	memberID string,
	models []string,
	members []types.Member,
	departments []types.Department,
	rules []types.RoutingRule,
	modelCatalog []types.ModelInfo,
	notInDeptMessage string,
) *string {
	if len(models) == 0 {
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
	allowed := ResolveDeptAllowedModels(member.DepartmentID, departments, rules, modelCatalog)
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, name := range allowed {
		allowedSet[name] = struct{}{}
	}
	for _, model := range models {
		if _, ok := allowedSet[model]; !ok {
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
	parentAllowed []string,
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
			ID: rule.ID, NodeID: rule.NodeID, NodeName: rule.NodeName,
			AllowedModels: append([]string{}, rule.AllowedModels...),
			Inherited:     rule.Inherited,
		}
		if rule.DefaultModel != nil {
			defaultModel := *rule.DefaultModel
			result[i].DefaultModel = &defaultModel
		}
		if rule.FallbackModel != nil {
			fallbackModel := *rule.FallbackModel
			result[i].FallbackModel = &fallbackModel
		}
	}
	return result
}

func shrinkChildRoutingRules(
	parentNodeID string,
	parentAllowed []string,
	rules []types.RoutingRule,
	parents map[string]*string,
) {
	for i := range rules {
		parentID := getParentDeptID(rules[i].NodeID, parents)
		if parentID == nil || *parentID != parentNodeID {
			continue
		}
		filtered := make([]string, 0)
		for _, model := range rules[i].AllowedModels {
			for _, allowed := range parentAllowed {
				if model == allowed {
					filtered = append(filtered, model)
					break
				}
			}
		}
		if len(filtered) == 0 && len(parentAllowed) > 0 {
			filtered = append([]string{}, parentAllowed...)
		}
		rules[i].AllowedModels = filtered
		shrinkChildRoutingRules(rules[i].NodeID, rules[i].AllowedModels, rules, parents)
	}
}
