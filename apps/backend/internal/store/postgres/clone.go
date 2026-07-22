package postgres

import (
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
)

func cloneProjects(items []types.Project) []types.Project {
	result := make([]types.Project, len(items))
	for i, project := range items {
		result[i] = types.Project{
			ID: project.ID, Name: project.Name, Budget: project.Budget, Consumed: project.Consumed,
			MemberIDs:         append([]uuid.UUID{}, project.MemberIDs...),
			OwnerDepartmentID: project.OwnerDepartmentID,
		}
		if len(project.MemberBudgets) > 0 {
			result[i].MemberBudgets = make(map[uuid.UUID]int64, len(project.MemberBudgets))
			for k, v := range project.MemberBudgets {
				result[i].MemberBudgets[k] = v
			}
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
			NotifyRoleIDs: append([]uuid.UUID{}, rule.NotifyRoleIDs...),
			Enabled:       rule.Enabled,
		}
	}
	return result
}

func cloneBudgetApprovals(items []types.BudgetApproval) []types.BudgetApproval {
	if len(items) == 0 {
		return nil
	}
	cloned := make([]types.BudgetApproval, len(items))
	for i, item := range items {
		cloned[i] = item
		if item.ResolvedAt != nil {
			s := *item.ResolvedAt
			cloned[i].ResolvedAt = &s
		}
		if item.RejectReason != nil {
			s := *item.RejectReason
			cloned[i].RejectReason = &s
		}
	}
	return cloned
}

func cloneProviderKeys(items []types.ProviderKey) []types.ProviderKey {
	result := make([]types.ProviderKey, len(items))
	for i, key := range items {
		cloned := types.ProviderKey{
			ID: key.ID, Provider: key.Provider, Name: key.Name, KeyPrefix: key.KeyPrefix,
			SecretKey: key.SecretKey, NewAPIChannelID: key.NewAPIChannelID,
			Status: key.Status, CreatedAt: key.CreatedAt, RotateEnabled: key.RotateEnabled,
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
		ID: key.ID, Name: key.Name, KeyPrefix: key.KeyPrefix, Scope: key.Scope, Status: key.Status,
		Budget: key.Budget, Consumed: key.Consumed, CreatedAt: key.CreatedAt,
		ModelWhitelist: append([]uuid.UUID{}, key.ModelWhitelist...),
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
	if key.ProjectID != nil {
		projectID := *key.ProjectID
		cloned.ProjectID = &projectID
	}
	if key.ProjectName != nil {
		projectName := *key.ProjectName
		cloned.ProjectName = &projectName
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
			Reason: approval.Reason, RequestedBudget: approval.RequestedBudget,
			RequestedModels: append([]uuid.UUID{}, approval.RequestedModels...),
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
		cloned := types.ModelInfo{
			ID: model.ID, CompanyID: model.CompanyID, Provider: model.Provider, Type: model.Type,
			Name: model.Name, Description: model.Description,
			MaxContext: model.MaxContext, Enabled: model.Enabled,
			Capabilities: append([]string{}, model.Capabilities...),
		}
		if model.Endpoint != nil {
			endpoint := *model.Endpoint
			cloned.Endpoint = &endpoint
		}
		result[i] = cloned
	}
	return result
}

func cloneMembers(items []types.Member) []types.Member {
	result := make([]types.Member, len(items))
	for i, member := range items {
		result[i] = types.Member{
			ID: member.ID, CompanyID: member.CompanyID, UserID: member.UserID, Alias: member.Alias, Avatar: member.Avatar,
			DepartmentID: member.DepartmentID, DepartmentName: member.DepartmentName,
			Status: member.Status, Roles: append([]string{}, member.Roles...), Source: member.Source,
			PersonalBudget: member.PersonalBudget,
		}
		if member.ExternalID != nil {
			externalID := *member.ExternalID
			result[i].ExternalID = &externalID
		}
	}
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

func cloneOrgNode(node types.OrgNode) types.OrgNode {
	cloned := types.OrgNode{
		ID: node.ID, Name: node.Name, MemberCount: node.MemberCount,
		Budget: node.Budget, Consumed: node.Consumed, Period: node.Period,
		RoutingInherited: node.RoutingInherited,
	}
	if node.ParentID != nil {
		parentID := *node.ParentID
		cloned.ParentID = &parentID
	}
	if node.ExternalID != nil {
		externalID := *node.ExternalID
		cloned.ExternalID = &externalID
	}
	if node.Source != nil {
		source := *node.Source
		cloned.Source = &source
	}
	if node.ManagerID != nil {
		managerID := *node.ManagerID
		cloned.ManagerID = &managerID
	}
	if node.ReservedPool != nil {
		reserved := *node.ReservedPool
		cloned.ReservedPool = &reserved
	}
	if node.DefaultModelID != nil {
		defaultModelID := *node.DefaultModelID
		cloned.DefaultModelID = &defaultModelID
	}
	if node.FallbackModelID != nil {
		fallbackModelID := *node.FallbackModelID
		cloned.FallbackModelID = &fallbackModelID
	}
	if len(node.Children) > 0 {
		cloned.Children = make([]types.OrgNode, len(node.Children))
		for i, child := range node.Children {
			cloned.Children[i] = cloneOrgNode(child)
		}
	}
	return cloned
}

func cloneOrgNodes(items []types.OrgNode) []types.OrgNode {
	result := make([]types.OrgNode, len(items))
	for i, node := range items {
		result[i] = cloneOrgNode(node)
	}
	return result
}
