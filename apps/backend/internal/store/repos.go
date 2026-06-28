package store

import "github.com/tokenjoy/backend/internal/domain/types"

type memoryOrgRepo struct{ store *Memory }

func (r *memoryOrgRepo) DataSourceStatus() types.DataSourceStatus {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return r.store.data.DataSourceStatus
}

func (r *memoryOrgRepo) SetDataSourceStatus(status types.DataSourceStatus) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.DataSourceStatus = status
	return nil
}

func (r *memoryOrgRepo) ImportFailures() []types.ImportFailure {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneImportFailures(r.store.data.ImportFailures)
}

func (r *memoryOrgRepo) SetImportFailures(failures []types.ImportFailure) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.ImportFailures = cloneImportFailures(failures)
	return nil
}

func (r *memoryOrgRepo) SyncConfig() types.SyncConfig {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return r.store.data.SyncConfig
}

func (r *memoryOrgRepo) SetSyncConfig(cfg types.SyncConfig) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.SyncConfig = cfg
	return nil
}

func (r *memoryOrgRepo) SyncLogs() []types.SyncLog {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneSyncLogs(r.store.data.SyncLogs)
}

func (r *memoryOrgRepo) AppendSyncLog(log types.SyncLog) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.SyncLogs = append([]types.SyncLog{log}, r.store.data.SyncLogs...)
	return nil
}

func (r *memoryOrgRepo) Departments() []types.Department {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneDepartments(r.store.data.Departments)
}

func (r *memoryOrgRepo) SetDepartments(departments []types.Department) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.Departments = cloneDepartments(departments)
	return nil
}

func (r *memoryOrgRepo) Members() []types.Member {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneMembers(r.store.data.Members)
}

func (r *memoryOrgRepo) SetMembers(members []types.Member) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.Members = cloneMembers(members)
	return nil
}

func (r *memoryOrgRepo) Roles() []types.Role {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneRoles(r.store.data.Roles)
}

func (r *memoryOrgRepo) SetRoles(roles []types.Role) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.Roles = cloneRoles(roles)
	return nil
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

func (r *memoryBudgetRepo) SetTree(tree []types.BudgetNode) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.BudgetTree = cloneBudgetTree(tree)
	return nil
}

func (r *memoryBudgetRepo) Groups() []types.BudgetGroup {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneBudgetGroups(r.store.data.BudgetGroups)
}

func (r *memoryBudgetRepo) SetGroups(groups []types.BudgetGroup) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.BudgetGroups = cloneBudgetGroups(groups)
	return nil
}

func (r *memoryBudgetRepo) OverrunPolicy() types.OverrunPolicyConfig {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return r.store.data.OverrunPolicy
}

func (r *memoryBudgetRepo) SetOverrunPolicy(policy types.OverrunPolicyConfig) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.OverrunPolicy = policy
	return nil
}

func (r *memoryBudgetRepo) AlertRules() []types.AlertRule {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneAlertRules(r.store.data.AlertRules)
}

func (r *memoryBudgetRepo) SetAlertRules(rules []types.AlertRule) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.AlertRules = cloneAlertRules(rules)
	return nil
}

func (r *memoryBudgetRepo) MemberQuotaPools() map[string]types.MemberQuotaPool {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneMemberQuotaPools(r.store.data.MemberQuotaPools)
}

func (r *memoryBudgetRepo) SetMemberQuotaPools(pools map[string]types.MemberQuotaPool) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.MemberQuotaPools = cloneMemberQuotaPools(pools)
	return nil
}

type memoryKeysRepo struct{ store *Memory }

func (r *memoryKeysRepo) ProviderKeys() []types.ProviderKey {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneProviderKeys(r.store.data.ProviderKeys)
}

func (r *memoryKeysRepo) SetProviderKeys(keys []types.ProviderKey) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.ProviderKeys = cloneProviderKeys(keys)
	return nil
}

func (r *memoryKeysRepo) PlatformKeys() []types.PlatformKey {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return clonePlatformKeys(r.store.data.PlatformKeys)
}

func (r *memoryKeysRepo) SetPlatformKeys(keys []types.PlatformKey) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.PlatformKeys = clonePlatformKeys(keys)
	return nil
}

func (r *memoryKeysRepo) Approvals() []types.KeyApproval {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneApprovals(r.store.data.Approvals)
}

func (r *memoryKeysRepo) SetApprovals(approvals []types.KeyApproval) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.Approvals = cloneApprovals(approvals)
	return nil
}

type memoryModelsRepo struct{ store *Memory }

func (r *memoryModelsRepo) Models() []types.ModelInfo {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneModels(r.store.data.Models)
}

func (r *memoryModelsRepo) SetModels(models []types.ModelInfo) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.Models = cloneModels(models)
	return nil
}

func (r *memoryModelsRepo) RoutingRules() []types.RoutingRule {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return cloneRoutingRules(r.store.data.RoutingRules)
}

func (r *memoryModelsRepo) SetRoutingRules(rules []types.RoutingRule) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.RoutingRules = cloneRoutingRules(rules)
	return nil
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

func (r *memoryAuditRepo) SetSettings(settings types.AuditSettings) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.AuditSettings = settings
	return nil
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
