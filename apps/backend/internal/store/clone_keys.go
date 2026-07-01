package store

import "github.com/tokenjoy/backend/internal/domain/types"

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

func CloneProviderKeys(items []types.ProviderKey) []types.ProviderKey {
	return cloneProviderKeys(items)
}

func ClonePlatformKeys(items []types.PlatformKey) []types.PlatformKey {
	return clonePlatformKeys(items)
}

func CloneApprovals(items []types.KeyApproval) []types.KeyApproval { return cloneApprovals(items) }
