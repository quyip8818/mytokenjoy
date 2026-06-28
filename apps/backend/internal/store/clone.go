package store

import "github.com/tokenjoy/backend/internal/domain/types"

func cloneImportFailures(items []types.ImportFailure) []types.ImportFailure {
	result := make([]types.ImportFailure, len(items))
	copy(result, items)
	return result
}

func cloneSyncLogs(items []types.SyncLog) []types.SyncLog {
	result := make([]types.SyncLog, len(items))
	copy(result, items)
	return result
}

func clonePermissions(items []types.Permission) []types.Permission {
	result := make([]types.Permission, len(items))
	copy(result, items)
	return result
}

func cloneRoles(items []types.Role) []types.Role {
	result := make([]types.Role, len(items))
	for i, role := range items {
		result[i] = types.Role{
			ID: role.ID, Name: role.Name, Type: role.Type,
			Permissions: append([]string{}, role.Permissions...),
			MemberCount: role.MemberCount,
		}
	}
	return result
}

func cloneMembers(items []types.Member) []types.Member {
	result := make([]types.Member, len(items))
	for i, member := range items {
		result[i] = types.Member{
			ID: member.ID, Name: member.Name, Phone: member.Phone, Email: member.Email,
			DepartmentID: member.DepartmentID, DepartmentName: member.DepartmentName,
			Status: member.Status, Roles: append([]string{}, member.Roles...), Source: member.Source,
		}
		if member.ExternalID != nil {
			externalID := *member.ExternalID
			result[i].ExternalID = &externalID
		}
	}
	return result
}

func cloneDepartments(items []types.Department) []types.Department {
	result := make([]types.Department, len(items))
	for i, dept := range items {
		result[i] = cloneDepartment(dept)
	}
	return result
}

func cloneDepartment(dept types.Department) types.Department {
	cloned := types.Department{ID: dept.ID, Name: dept.Name, MemberCount: dept.MemberCount}
	if dept.ParentID != nil {
		parentID := *dept.ParentID
		cloned.ParentID = &parentID
	}
	if dept.ExternalID != nil {
		externalID := *dept.ExternalID
		cloned.ExternalID = &externalID
	}
	if dept.Source != nil {
		source := *dept.Source
		cloned.Source = &source
	}
	if dept.ManagerID != nil {
		managerID := *dept.ManagerID
		cloned.ManagerID = &managerID
	}
	if len(dept.Children) > 0 {
		cloned.Children = make([]types.Department, len(dept.Children))
		for i, child := range dept.Children {
			cloned.Children[i] = cloneDepartment(child)
		}
	}
	return cloned
}

func cloneBudgetNode(node types.BudgetNode) types.BudgetNode {
	cloned := types.BudgetNode{
		ID: node.ID, Name: node.Name, Budget: node.Budget,
		Consumed: node.Consumed, Period: node.Period,
	}
	if node.ParentID != nil {
		parentID := *node.ParentID
		cloned.ParentID = &parentID
	}
	if node.ReservedPool != nil {
		reserved := *node.ReservedPool
		cloned.ReservedPool = &reserved
	}
	if len(node.Children) > 0 {
		cloned.Children = make([]types.BudgetNode, len(node.Children))
		for i, child := range node.Children {
			cloned.Children[i] = cloneBudgetNode(child)
		}
	}
	return cloned
}

func cloneBudgetTree(items []types.BudgetNode) []types.BudgetNode {
	result := make([]types.BudgetNode, len(items))
	for i, node := range items {
		result[i] = cloneBudgetNode(node)
	}
	return result
}

func cloneBudgetGroups(items []types.BudgetGroup) []types.BudgetGroup {
	result := make([]types.BudgetGroup, len(items))
	for i, group := range items {
		result[i] = types.BudgetGroup{
			ID: group.ID, Name: group.Name, Budget: group.Budget, Consumed: group.Consumed,
			MemberIDs:     append([]string{}, group.MemberIDs...),
			DepartmentIDs: append([]string{}, group.DepartmentIDs...),
		}
	}
	return result
}

func cloneAlertRules(items []types.AlertRule) []types.AlertRule {
	result := make([]types.AlertRule, len(items))
	for i, rule := range items {
		result[i] = types.AlertRule{
			ID: rule.ID, NodeID: rule.NodeID, NodeName: rule.NodeName,
			Thresholds:    append([]int{}, rule.Thresholds...),
			NotifyRoleIDs: append([]string{}, rule.NotifyRoleIDs...),
			Enabled:       rule.Enabled,
		}
	}
	return result
}

func cloneMemberQuotaPools(pools map[string]types.MemberQuotaPool) map[string]types.MemberQuotaPool {
	if pools == nil {
		return map[string]types.MemberQuotaPool{}
	}
	result := make(map[string]types.MemberQuotaPool, len(pools))
	for key, pool := range pools {
		result[key] = pool
	}
	return result
}

func cloneProviderKeys(items []types.ProviderKey) []types.ProviderKey {
	result := make([]types.ProviderKey, len(items))
	for i, key := range items {
		cloned := types.ProviderKey{
			ID: key.ID, Provider: key.Provider, Name: key.Name, KeyPrefix: key.KeyPrefix,
			Status: key.Status, CreatedAt: key.CreatedAt, RotateEnabled: key.RotateEnabled,
		}
		if key.Balance != nil {
			balance := *key.Balance
			cloned.Balance = &balance
		}
		if key.LastUsed != nil {
			lastUsed := *key.LastUsed
			cloned.LastUsed = &lastUsed
		}
		result[i] = cloned
	}
	return result
}

func clonePlatformKeys(items []types.PlatformKey) []types.PlatformKey {
	result := make([]types.PlatformKey, len(items))
	for i, key := range items {
		result[i] = clonePlatformKey(key)
	}
	return result
}

func clonePlatformKey(key types.PlatformKey) types.PlatformKey {
	cloned := types.PlatformKey{
		ID: key.ID, Name: key.Name, KeyPrefix: key.KeyPrefix, Status: key.Status,
		Quota: key.Quota, Used: key.Used, CreatedAt: key.CreatedAt,
		ModelWhitelist: append([]string{}, key.ModelWhitelist...),
	}
	if key.FullKey != nil {
		fullKey := *key.FullKey
		cloned.FullKey = &fullKey
	}
	if key.MemberID != nil {
		memberID := *key.MemberID
		cloned.MemberID = &memberID
	}
	if key.MemberName != nil {
		memberName := *key.MemberName
		cloned.MemberName = &memberName
	}
	if key.AppName != nil {
		appName := *key.AppName
		cloned.AppName = &appName
	}
	if key.BudgetGroupID != nil {
		budgetGroupID := *key.BudgetGroupID
		cloned.BudgetGroupID = &budgetGroupID
	}
	if key.BudgetGroupName != nil {
		budgetGroupName := *key.BudgetGroupName
		cloned.BudgetGroupName = &budgetGroupName
	}
	if key.ExpiresAt != nil {
		expiresAt := *key.ExpiresAt
		cloned.ExpiresAt = &expiresAt
	}
	return cloned
}

func cloneApprovals(items []types.KeyApproval) []types.KeyApproval {
	result := make([]types.KeyApproval, len(items))
	for i, approval := range items {
		cloned := types.KeyApproval{
			ID: approval.ID, Type: approval.Type, Applicant: approval.Applicant,
			ApplicantID: approval.ApplicantID, Department: approval.Department,
			Reason: approval.Reason, RequestedQuota: approval.RequestedQuota,
			RequestedModels: append([]string{}, approval.RequestedModels...),
			Status:          approval.Status, CreatedAt: approval.CreatedAt,
		}
		if approval.Approver != nil {
			approver := *approval.Approver
			cloned.Approver = &approver
		}
		if approval.RejectReason != nil {
			rejectReason := *approval.RejectReason
			cloned.RejectReason = &rejectReason
		}
		if approval.ResolvedAt != nil {
			resolvedAt := *approval.ResolvedAt
			cloned.ResolvedAt = &resolvedAt
		}
		result[i] = cloned
	}
	return result
}

func cloneModels(items []types.ModelInfo) []types.ModelInfo {
	result := make([]types.ModelInfo, len(items))
	for i, model := range items {
		result[i] = types.ModelInfo{
			ID: model.ID, Provider: model.Provider, Name: model.Name,
			DisplayName: model.DisplayName, InputPrice: model.InputPrice,
			OutputPrice: model.OutputPrice, MaxContext: model.MaxContext, Enabled: model.Enabled,
			Capabilities: append([]string{}, model.Capabilities...),
		}
	}
	return result
}

func cloneRoutingRules(items []types.RoutingRule) []types.RoutingRule {
	result := make([]types.RoutingRule, len(items))
	for i, rule := range items {
		cloned := types.RoutingRule{
			ID: rule.ID, NodeID: rule.NodeID, NodeName: rule.NodeName,
			AllowedModels: append([]string{}, rule.AllowedModels...),
			Inherited:     rule.Inherited,
		}
		if rule.DefaultModel != nil {
			defaultModel := *rule.DefaultModel
			cloned.DefaultModel = &defaultModel
		}
		if rule.FallbackModel != nil {
			fallbackModel := *rule.FallbackModel
			cloned.FallbackModel = &fallbackModel
		}
		result[i] = cloned
	}
	return result
}

func cloneModelUsage(items []types.ModelUsage) []types.ModelUsage {
	result := make([]types.ModelUsage, len(items))
	copy(result, items)
	return result
}

func cloneTeamUsage(items []types.TeamUsage) []types.TeamUsage {
	result := make([]types.TeamUsage, len(items))
	copy(result, items)
	return result
}

func cloneOperationLogs(items []types.OperationLog) []types.OperationLog {
	result := make([]types.OperationLog, len(items))
	copy(result, items)
	return result
}

func cloneCallLogs(items []types.CallLog) []types.CallLog {
	result := make([]types.CallLog, len(items))
	copy(result, items)
	return result
}
