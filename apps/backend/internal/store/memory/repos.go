package memory

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type memoryOrgRepo struct{ store *Store }

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
	return store.CloneImportFailures(r.store.data.ImportFailures)
}

func (r *memoryOrgRepo) SetImportFailures(failures []types.ImportFailure) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.ImportFailures = store.CloneImportFailures(failures)
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
	return store.CloneSyncLogs(r.store.data.SyncLogs)
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
	return store.CloneDepartments(r.store.data.Departments)
}

func (r *memoryOrgRepo) SetDepartments(departments []types.Department) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.Departments = store.CloneDepartments(departments)
	return nil
}

func (r *memoryOrgRepo) Members() []types.Member {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneMembers(r.store.data.Members)
}

func (r *memoryOrgRepo) SetMembers(members []types.Member) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.Members = store.CloneMembers(members)
	return nil
}

func (r *memoryOrgRepo) Roles() []types.Role {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneRoles(r.store.data.Roles)
}

func (r *memoryOrgRepo) SetRoles(roles []types.Role) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.Roles = store.CloneRoles(roles)
	return nil
}

func (r *memoryOrgRepo) Permissions() []types.Permission {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.ClonePermissions(r.store.data.Permissions)
}

type memoryBudgetRepo struct{ store *Store }

func (r *memoryBudgetRepo) Tree() []types.BudgetNode {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneBudgetTree(r.store.data.BudgetTree)
}

func (r *memoryBudgetRepo) SetTree(tree []types.BudgetNode) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.BudgetTree = store.CloneBudgetTree(tree)
	return nil
}

func (r *memoryBudgetRepo) Groups() []types.BudgetGroup {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneBudgetGroups(r.store.data.BudgetGroups)
}

func (r *memoryBudgetRepo) SetGroups(groups []types.BudgetGroup) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.BudgetGroups = store.CloneBudgetGroups(groups)
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
	return store.CloneAlertRules(r.store.data.AlertRules)
}

func (r *memoryBudgetRepo) SetAlertRules(rules []types.AlertRule) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.AlertRules = store.CloneAlertRules(rules)
	return nil
}

func (r *memoryBudgetRepo) MemberQuotaPools() map[string]types.MemberQuotaPool {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneMemberQuotaPools(r.store.data.MemberQuotaPools)
}

func (r *memoryBudgetRepo) SetMemberQuotaPools(pools map[string]types.MemberQuotaPool) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.MemberQuotaPools = store.CloneMemberQuotaPools(pools)
	return nil
}

type memoryKeysRepo struct{ store *Store }

func (r *memoryKeysRepo) ProviderKeys() []types.ProviderKey {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneProviderKeys(r.store.data.ProviderKeys)
}

func (r *memoryKeysRepo) SetProviderKeys(keys []types.ProviderKey) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.ProviderKeys = store.CloneProviderKeys(keys)
	return nil
}

func (r *memoryKeysRepo) PlatformKeys() []types.PlatformKey {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.ClonePlatformKeys(r.store.data.PlatformKeys)
}

func (r *memoryKeysRepo) SetPlatformKeys(keys []types.PlatformKey) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.PlatformKeys = store.ClonePlatformKeys(keys)
	return nil
}

func (r *memoryKeysRepo) Approvals() []types.KeyApproval {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneApprovals(r.store.data.Approvals)
}

func (r *memoryKeysRepo) SetApprovals(approvals []types.KeyApproval) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.Approvals = store.CloneApprovals(approvals)
	return nil
}

type memoryModelsRepo struct{ store *Store }

func (r *memoryModelsRepo) Models() []types.ModelInfo {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneModels(r.store.data.Models)
}

func (r *memoryModelsRepo) SetModels(models []types.ModelInfo) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.Models = store.CloneModels(models)
	return nil
}

func (r *memoryModelsRepo) RoutingRules() []types.RoutingRule {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneRoutingRules(r.store.data.RoutingRules)
}

func (r *memoryModelsRepo) SetRoutingRules(rules []types.RoutingRule) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	r.store.data.RoutingRules = store.CloneRoutingRules(rules)
	return nil
}

type memoryAuditRepo struct{ store *Store }

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
	return store.CloneOperationLogs(r.store.data.OperationLogs)
}

func (r *memoryAuditRepo) CallLogs() []types.CallLog {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneCallLogs(r.store.data.CallLogs)
}
