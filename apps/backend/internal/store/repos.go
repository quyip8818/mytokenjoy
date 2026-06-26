package store

import "github.com/tokenjoy/backend/internal/domain/types"

type memoryOrgRepo struct{ store *Memory }

func (r *memoryOrgRepo) DataSourceStatus() types.DataSourceStatus {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return r.store.data.DataSourceStatus
}

func (r *memoryOrgRepo) SetDataSourceStatus(status types.DataSourceStatus) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.DataSourceStatus = status
}

func (r *memoryOrgRepo) ImportFailures() []types.ImportFailure {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneImportFailures(r.store.data.ImportFailures)
}

func (r *memoryOrgRepo) SyncConfig() types.SyncConfig {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return r.store.data.SyncConfig
}

func (r *memoryOrgRepo) SetSyncConfig(cfg types.SyncConfig) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.SyncConfig = cfg
}

func (r *memoryOrgRepo) SyncLogs() []types.SyncLog {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneSyncLogs(r.store.data.SyncLogs)
}

func (r *memoryOrgRepo) Departments() []types.Department {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneDepartments(r.store.data.Departments)
}

func (r *memoryOrgRepo) SetDepartments(departments []types.Department) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.Departments = cloneDepartments(departments)
}

func (r *memoryOrgRepo) Members() []types.Member {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneMembers(r.store.data.Members)
}

func (r *memoryOrgRepo) SetMembers(members []types.Member) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.Members = cloneMembers(members)
}

func (r *memoryOrgRepo) Roles() []types.Role {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneRoles(r.store.data.Roles)
}

func (r *memoryOrgRepo) SetRoles(roles []types.Role) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.Roles = cloneRoles(roles)
}

func (r *memoryOrgRepo) Permissions() []types.Permission {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return clonePermissions(r.store.data.Permissions)
}

type memoryBudgetRepo struct{ store *Memory }

func (r *memoryBudgetRepo) Tree() []types.BudgetNode {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneBudgetTree(r.store.data.BudgetTree)
}

func (r *memoryBudgetRepo) SetTree(tree []types.BudgetNode) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.BudgetTree = cloneBudgetTree(tree)
}

func (r *memoryBudgetRepo) Groups() []types.BudgetGroup {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneBudgetGroups(r.store.data.BudgetGroups)
}

func (r *memoryBudgetRepo) SetGroups(groups []types.BudgetGroup) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.BudgetGroups = cloneBudgetGroups(groups)
}

func (r *memoryBudgetRepo) OverrunPolicy() types.OverrunPolicyConfig {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return r.store.data.OverrunPolicy
}

func (r *memoryBudgetRepo) SetOverrunPolicy(policy types.OverrunPolicyConfig) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.OverrunPolicy = policy
}

func (r *memoryBudgetRepo) AlertRules() []types.AlertRule {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneAlertRules(r.store.data.AlertRules)
}

func (r *memoryBudgetRepo) SetAlertRules(rules []types.AlertRule) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.AlertRules = cloneAlertRules(rules)
}

func (r *memoryBudgetRepo) MemberQuotaPools() map[string]types.MemberQuotaPool {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneMemberQuotaPools(r.store.data.MemberQuotaPools)
}

func (r *memoryBudgetRepo) SetMemberQuotaPools(pools map[string]types.MemberQuotaPool) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.MemberQuotaPools = cloneMemberQuotaPools(pools)
}

type memoryKeysRepo struct{ store *Memory }

func (r *memoryKeysRepo) ProviderKeys() []types.ProviderKey {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneProviderKeys(r.store.data.ProviderKeys)
}

func (r *memoryKeysRepo) SetProviderKeys(keys []types.ProviderKey) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.ProviderKeys = cloneProviderKeys(keys)
}

func (r *memoryKeysRepo) PlatformKeys() []types.PlatformKey {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return clonePlatformKeys(r.store.data.PlatformKeys)
}

func (r *memoryKeysRepo) SetPlatformKeys(keys []types.PlatformKey) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.PlatformKeys = clonePlatformKeys(keys)
}

func (r *memoryKeysRepo) Approvals() []types.KeyApproval {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneApprovals(r.store.data.Approvals)
}

func (r *memoryKeysRepo) SetApprovals(approvals []types.KeyApproval) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.Approvals = cloneApprovals(approvals)
}

type memoryModelsRepo struct{ store *Memory }

func (r *memoryModelsRepo) Models() []types.ModelInfo {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneModels(r.store.data.Models)
}

func (r *memoryModelsRepo) SetModels(models []types.ModelInfo) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.Models = cloneModels(models)
}

func (r *memoryModelsRepo) RoutingRules() []types.RoutingRule {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneRoutingRules(r.store.data.RoutingRules)
}

func (r *memoryModelsRepo) SetRoutingRules(rules []types.RoutingRule) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.RoutingRules = cloneRoutingRules(rules)
}

type memoryDashboardRepo struct{ store *Memory }

func (r *memoryDashboardRepo) ModelUsage() []types.ModelUsage {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneModelUsage(r.store.data.ModelUsage)
}

func (r *memoryDashboardRepo) TeamUsage() []types.TeamUsage {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneTeamUsage(r.store.data.TeamUsage)
}

type memoryAuditRepo struct{ store *Memory }

func (r *memoryAuditRepo) Settings() types.AuditSettings {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return r.store.data.AuditSettings
}

func (r *memoryAuditRepo) SetSettings(settings types.AuditSettings) {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.AuditSettings = settings
}

func (r *memoryAuditRepo) OperationLogs() []types.OperationLog {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneOperationLogs(r.store.data.OperationLogs)
}

func (r *memoryAuditRepo) CallLogs() []types.CallLog {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneCallLogs(r.store.data.CallLogs)
}
