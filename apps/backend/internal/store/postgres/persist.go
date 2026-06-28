package postgres

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type persistOrgRepo struct {
	inner store.OrgRepository
	store *Store
}

func (r *persistOrgRepo) DataSourceStatus() types.DataSourceStatus { return r.inner.DataSourceStatus() }
func (r *persistOrgRepo) SetDataSourceStatus(status types.DataSourceStatus) error {
	if err := r.inner.SetDataSourceStatus(status); err != nil {
		return err
	}
	return r.store.persistSnapshot()
}
func (r *persistOrgRepo) ImportFailures() []types.ImportFailure { return r.inner.ImportFailures() }
func (r *persistOrgRepo) SetImportFailures(failures []types.ImportFailure) error {
	if err := r.inner.SetImportFailures(failures); err != nil {
		return err
	}
	return r.store.persistSnapshot()
}
func (r *persistOrgRepo) SyncConfig() types.SyncConfig { return r.inner.SyncConfig() }
func (r *persistOrgRepo) SetSyncConfig(cfg types.SyncConfig) error {
	if err := r.inner.SetSyncConfig(cfg); err != nil {
		return err
	}
	return r.store.persistSnapshot()
}
func (r *persistOrgRepo) SyncLogs() []types.SyncLog { return r.inner.SyncLogs() }
func (r *persistOrgRepo) AppendSyncLog(log types.SyncLog) error {
	if err := r.inner.AppendSyncLog(log); err != nil {
		return err
	}
	return r.store.persistSnapshot()
}
func (r *persistOrgRepo) Departments() []types.Department {
	return r.inner.Departments()
}
func (r *persistOrgRepo) SetDepartments(departments []types.Department) error {
	if err := r.inner.SetDepartments(departments); err != nil {
		return err
	}
	return r.store.persistSnapshot()
}
func (r *persistOrgRepo) Members() []types.Member { return r.inner.Members() }
func (r *persistOrgRepo) SetMembers(members []types.Member) error {
	if err := r.inner.SetMembers(members); err != nil {
		return err
	}
	return r.store.persistSnapshot()
}
func (r *persistOrgRepo) Roles() []types.Role { return r.inner.Roles() }
func (r *persistOrgRepo) SetRoles(roles []types.Role) error {
	if err := r.inner.SetRoles(roles); err != nil {
		return err
	}
	return r.store.persistSnapshot()
}
func (r *persistOrgRepo) Permissions() []types.Permission { return r.inner.Permissions() }

type persistBudgetRepo struct {
	inner store.BudgetRepository
	store *Store
}

func (r *persistBudgetRepo) Tree() []types.BudgetNode { return r.inner.Tree() }
func (r *persistBudgetRepo) SetTree(tree []types.BudgetNode) error {
	if err := r.inner.SetTree(tree); err != nil {
		return err
	}
	return r.store.persistSnapshot()
}
func (r *persistBudgetRepo) Groups() []types.BudgetGroup { return r.inner.Groups() }
func (r *persistBudgetRepo) SetGroups(groups []types.BudgetGroup) error {
	if err := r.inner.SetGroups(groups); err != nil {
		return err
	}
	return r.store.persistSnapshot()
}
func (r *persistBudgetRepo) OverrunPolicy() types.OverrunPolicyConfig { return r.inner.OverrunPolicy() }
func (r *persistBudgetRepo) SetOverrunPolicy(policy types.OverrunPolicyConfig) error {
	if err := r.inner.SetOverrunPolicy(policy); err != nil {
		return err
	}
	return r.store.persistSnapshot()
}
func (r *persistBudgetRepo) AlertRules() []types.AlertRule { return r.inner.AlertRules() }
func (r *persistBudgetRepo) SetAlertRules(rules []types.AlertRule) error {
	if err := r.inner.SetAlertRules(rules); err != nil {
		return err
	}
	return r.store.persistSnapshot()
}
func (r *persistBudgetRepo) MemberQuotaPools() map[string]types.MemberQuotaPool {
	return r.inner.MemberQuotaPools()
}
func (r *persistBudgetRepo) SetMemberQuotaPools(pools map[string]types.MemberQuotaPool) error {
	if err := r.inner.SetMemberQuotaPools(pools); err != nil {
		return err
	}
	return r.store.persistSnapshot()
}

type persistKeysRepo struct {
	inner store.KeysRepository
	store *Store
}

func (r *persistKeysRepo) ProviderKeys() []types.ProviderKey { return r.inner.ProviderKeys() }
func (r *persistKeysRepo) SetProviderKeys(keys []types.ProviderKey) error {
	if err := r.inner.SetProviderKeys(keys); err != nil {
		return err
	}
	return r.store.persistSnapshot()
}
func (r *persistKeysRepo) PlatformKeys() []types.PlatformKey { return r.inner.PlatformKeys() }
func (r *persistKeysRepo) SetPlatformKeys(keys []types.PlatformKey) error {
	if err := r.inner.SetPlatformKeys(keys); err != nil {
		return err
	}
	return r.store.persistSnapshot()
}
func (r *persistKeysRepo) Approvals() []types.KeyApproval { return r.inner.Approvals() }
func (r *persistKeysRepo) SetApprovals(approvals []types.KeyApproval) error {
	if err := r.inner.SetApprovals(approvals); err != nil {
		return err
	}
	return r.store.persistSnapshot()
}

type persistModelsRepo struct {
	inner store.ModelsRepository
	store *Store
}

func (r *persistModelsRepo) Models() []types.ModelInfo { return r.inner.Models() }
func (r *persistModelsRepo) SetModels(models []types.ModelInfo) error {
	if err := r.inner.SetModels(models); err != nil {
		return err
	}
	return r.store.persistSnapshot()
}
func (r *persistModelsRepo) RoutingRules() []types.RoutingRule { return r.inner.RoutingRules() }
func (r *persistModelsRepo) SetRoutingRules(rules []types.RoutingRule) error {
	if err := r.inner.SetRoutingRules(rules); err != nil {
		return err
	}
	return r.store.persistSnapshot()
}

type persistAuditRepo struct {
	inner store.AuditRepository
	store *Store
}

func (r *persistAuditRepo) Settings() types.AuditSettings { return r.inner.Settings() }
func (r *persistAuditRepo) SetSettings(settings types.AuditSettings) error {
	if err := r.inner.SetSettings(settings); err != nil {
		return err
	}
	return r.store.persistSnapshot()
}
func (r *persistAuditRepo) OperationLogs() []types.OperationLog { return r.inner.OperationLogs() }
func (r *persistAuditRepo) CallLogs() []types.CallLog           { return r.inner.CallLogs() }

type deferredOrgRepo struct {
	inner store.OrgRepository
}

func (r *deferredOrgRepo) DataSourceStatus() types.DataSourceStatus {
	return r.inner.DataSourceStatus()
}
func (r *deferredOrgRepo) SetDataSourceStatus(status types.DataSourceStatus) error {
	return r.inner.SetDataSourceStatus(status)
}
func (r *deferredOrgRepo) ImportFailures() []types.ImportFailure { return r.inner.ImportFailures() }
func (r *deferredOrgRepo) SetImportFailures(failures []types.ImportFailure) error {
	return r.inner.SetImportFailures(failures)
}
func (r *deferredOrgRepo) SyncConfig() types.SyncConfig { return r.inner.SyncConfig() }
func (r *deferredOrgRepo) SetSyncConfig(cfg types.SyncConfig) error {
	return r.inner.SetSyncConfig(cfg)
}
func (r *deferredOrgRepo) SyncLogs() []types.SyncLog { return r.inner.SyncLogs() }
func (r *deferredOrgRepo) AppendSyncLog(log types.SyncLog) error {
	return r.inner.AppendSyncLog(log)
}
func (r *deferredOrgRepo) Departments() []types.Department {
	return r.inner.Departments()
}
func (r *deferredOrgRepo) SetDepartments(departments []types.Department) error {
	return r.inner.SetDepartments(departments)
}
func (r *deferredOrgRepo) Members() []types.Member { return r.inner.Members() }
func (r *deferredOrgRepo) SetMembers(members []types.Member) error {
	return r.inner.SetMembers(members)
}
func (r *deferredOrgRepo) Roles() []types.Role { return r.inner.Roles() }
func (r *deferredOrgRepo) SetRoles(roles []types.Role) error {
	return r.inner.SetRoles(roles)
}
func (r *deferredOrgRepo) Permissions() []types.Permission { return r.inner.Permissions() }

type deferredBudgetRepo struct {
	inner store.BudgetRepository
}

func (r *deferredBudgetRepo) Tree() []types.BudgetNode { return r.inner.Tree() }
func (r *deferredBudgetRepo) SetTree(tree []types.BudgetNode) error {
	return r.inner.SetTree(tree)
}
func (r *deferredBudgetRepo) Groups() []types.BudgetGroup { return r.inner.Groups() }
func (r *deferredBudgetRepo) SetGroups(groups []types.BudgetGroup) error {
	return r.inner.SetGroups(groups)
}
func (r *deferredBudgetRepo) OverrunPolicy() types.OverrunPolicyConfig {
	return r.inner.OverrunPolicy()
}
func (r *deferredBudgetRepo) SetOverrunPolicy(policy types.OverrunPolicyConfig) error {
	return r.inner.SetOverrunPolicy(policy)
}
func (r *deferredBudgetRepo) AlertRules() []types.AlertRule { return r.inner.AlertRules() }
func (r *deferredBudgetRepo) SetAlertRules(rules []types.AlertRule) error {
	return r.inner.SetAlertRules(rules)
}
func (r *deferredBudgetRepo) MemberQuotaPools() map[string]types.MemberQuotaPool {
	return r.inner.MemberQuotaPools()
}
func (r *deferredBudgetRepo) SetMemberQuotaPools(pools map[string]types.MemberQuotaPool) error {
	return r.inner.SetMemberQuotaPools(pools)
}

type deferredKeysRepo struct {
	inner store.KeysRepository
}

func (r *deferredKeysRepo) ProviderKeys() []types.ProviderKey { return r.inner.ProviderKeys() }
func (r *deferredKeysRepo) SetProviderKeys(keys []types.ProviderKey) error {
	return r.inner.SetProviderKeys(keys)
}
func (r *deferredKeysRepo) PlatformKeys() []types.PlatformKey { return r.inner.PlatformKeys() }
func (r *deferredKeysRepo) SetPlatformKeys(keys []types.PlatformKey) error {
	return r.inner.SetPlatformKeys(keys)
}
func (r *deferredKeysRepo) Approvals() []types.KeyApproval { return r.inner.Approvals() }
func (r *deferredKeysRepo) SetApprovals(approvals []types.KeyApproval) error {
	return r.inner.SetApprovals(approvals)
}

type deferredModelsRepo struct {
	inner store.ModelsRepository
}

func (r *deferredModelsRepo) Models() []types.ModelInfo { return r.inner.Models() }
func (r *deferredModelsRepo) SetModels(models []types.ModelInfo) error {
	return r.inner.SetModels(models)
}
func (r *deferredModelsRepo) RoutingRules() []types.RoutingRule { return r.inner.RoutingRules() }
func (r *deferredModelsRepo) SetRoutingRules(rules []types.RoutingRule) error {
	return r.inner.SetRoutingRules(rules)
}

type deferredAuditRepo struct {
	inner store.AuditRepository
}

func (r *deferredAuditRepo) Settings() types.AuditSettings { return r.inner.Settings() }
func (r *deferredAuditRepo) SetSettings(settings types.AuditSettings) error {
	return r.inner.SetSettings(settings)
}
func (r *deferredAuditRepo) OperationLogs() []types.OperationLog { return r.inner.OperationLogs() }
func (r *deferredAuditRepo) CallLogs() []types.CallLog           { return r.inner.CallLogs() }
