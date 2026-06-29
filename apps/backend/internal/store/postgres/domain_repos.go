package postgres

import (
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type pgOrgRepo struct {
	backend shardBackend
}

func (r *pgOrgRepo) read() (store.OrgShardData, error) {
	raw, err := r.backend.load(store.ShardOrg)
	if err != nil {
		return store.OrgShardData{}, err
	}
	return store.ParseOrgShard(raw)
}

func (r *pgOrgRepo) write(data store.OrgShardData) error {
	raw, err := store.MarshalOrgShard(data)
	if err != nil {
		return err
	}
	return r.backend.save(store.ShardOrg, raw)
}

func (r *pgOrgRepo) DataSourceStatus() types.DataSourceStatus {
	data, err := r.read()
	if err != nil {
		return types.DataSourceStatus{}
	}
	return data.DataSourceStatus
}

func (r *pgOrgRepo) SetDataSourceStatus(status types.DataSourceStatus) error {
	data, err := r.read()
	if err != nil {
		return err
	}
	data.DataSourceStatus = status
	return r.write(data)
}

func (r *pgOrgRepo) ImportFailures() []types.ImportFailure {
	data, err := r.read()
	if err != nil {
		return nil
	}
	return store.CloneImportFailures(data.ImportFailures)
}

func (r *pgOrgRepo) SetImportFailures(failures []types.ImportFailure) error {
	data, err := r.read()
	if err != nil {
		return err
	}
	data.ImportFailures = store.CloneImportFailures(failures)
	return r.write(data)
}

func (r *pgOrgRepo) SyncConfig() types.SyncConfig {
	data, err := r.read()
	if err != nil {
		return types.SyncConfig{}
	}
	return data.SyncConfig
}

func (r *pgOrgRepo) SetSyncConfig(cfg types.SyncConfig) error {
	data, err := r.read()
	if err != nil {
		return err
	}
	data.SyncConfig = cfg
	return r.write(data)
}

func (r *pgOrgRepo) SyncLogs() []types.SyncLog {
	data, err := r.read()
	if err != nil {
		return nil
	}
	return store.CloneSyncLogs(data.SyncLogs)
}

func (r *pgOrgRepo) AppendSyncLog(log types.SyncLog) error {
	data, err := r.read()
	if err != nil {
		return err
	}
	data.SyncLogs = append([]types.SyncLog{log}, data.SyncLogs...)
	return r.write(data)
}

func (r *pgOrgRepo) Departments() []types.Department {
	data, err := r.read()
	if err != nil {
		return nil
	}
	return store.CloneDepartments(data.Departments)
}

func (r *pgOrgRepo) SetDepartments(departments []types.Department) error {
	data, err := r.read()
	if err != nil {
		return err
	}
	data.Departments = store.CloneDepartments(departments)
	return r.write(data)
}

func (r *pgOrgRepo) Members() []types.Member {
	data, err := r.read()
	if err != nil {
		return nil
	}
	return store.CloneMembers(data.Members)
}

func (r *pgOrgRepo) SetMembers(members []types.Member) error {
	data, err := r.read()
	if err != nil {
		return err
	}
	data.Members = store.CloneMembers(members)
	return r.write(data)
}

func (r *pgOrgRepo) Roles() []types.Role {
	data, err := r.read()
	if err != nil {
		return nil
	}
	return store.CloneRoles(data.Roles)
}

func (r *pgOrgRepo) SetRoles(roles []types.Role) error {
	data, err := r.read()
	if err != nil {
		return err
	}
	data.Roles = store.CloneRoles(roles)
	return r.write(data)
}

func (r *pgOrgRepo) Permissions() []types.Permission {
	data, err := r.read()
	if err != nil {
		return nil
	}
	return store.ClonePermissions(data.Permissions)
}

type pgBudgetRepo struct {
	backend shardBackend
}

func (r *pgBudgetRepo) read() (store.BudgetShardData, error) {
	raw, err := r.backend.load(store.ShardBudget)
	if err != nil {
		return store.BudgetShardData{}, err
	}
	return store.ParseBudgetShard(raw)
}

func (r *pgBudgetRepo) write(data store.BudgetShardData) error {
	raw, err := store.MarshalBudgetShard(data)
	if err != nil {
		return err
	}
	return r.backend.save(store.ShardBudget, raw)
}

func (r *pgBudgetRepo) Tree() []types.BudgetNode {
	data, err := r.read()
	if err != nil {
		return nil
	}
	return store.CloneBudgetTree(data.BudgetTree)
}

func (r *pgBudgetRepo) SetTree(tree []types.BudgetNode) error {
	data, err := r.read()
	if err != nil {
		return err
	}
	data.BudgetTree = store.CloneBudgetTree(tree)
	return r.write(data)
}

func (r *pgBudgetRepo) Groups() []types.BudgetGroup {
	data, err := r.read()
	if err != nil {
		return nil
	}
	return store.CloneBudgetGroups(data.BudgetGroups)
}

func (r *pgBudgetRepo) SetGroups(groups []types.BudgetGroup) error {
	data, err := r.read()
	if err != nil {
		return err
	}
	data.BudgetGroups = store.CloneBudgetGroups(groups)
	return r.write(data)
}

func (r *pgBudgetRepo) OverrunPolicy() types.OverrunPolicyConfig {
	data, err := r.read()
	if err != nil {
		return types.OverrunPolicyConfig{}
	}
	return data.OverrunPolicy
}

func (r *pgBudgetRepo) SetOverrunPolicy(policy types.OverrunPolicyConfig) error {
	data, err := r.read()
	if err != nil {
		return err
	}
	data.OverrunPolicy = policy
	return r.write(data)
}

func (r *pgBudgetRepo) AlertRules() []types.AlertRule {
	data, err := r.read()
	if err != nil {
		return nil
	}
	return store.CloneAlertRules(data.AlertRules)
}

func (r *pgBudgetRepo) SetAlertRules(rules []types.AlertRule) error {
	data, err := r.read()
	if err != nil {
		return err
	}
	data.AlertRules = store.CloneAlertRules(rules)
	return r.write(data)
}

func (r *pgBudgetRepo) MemberQuotaPools() map[string]types.MemberQuotaPool {
	data, err := r.read()
	if err != nil {
		return map[string]types.MemberQuotaPool{}
	}
	return store.CloneMemberQuotaPools(data.MemberQuotaPools)
}

func (r *pgBudgetRepo) SetMemberQuotaPools(pools map[string]types.MemberQuotaPool) error {
	data, err := r.read()
	if err != nil {
		return err
	}
	data.MemberQuotaPools = store.CloneMemberQuotaPools(pools)
	return r.write(data)
}

type pgKeysRepo struct {
	backend shardBackend
}

func (r *pgKeysRepo) read() (store.KeysShardData, error) {
	raw, err := r.backend.load(store.ShardKeys)
	if err != nil {
		return store.KeysShardData{}, err
	}
	return store.ParseKeysShard(raw)
}

func (r *pgKeysRepo) write(data store.KeysShardData) error {
	raw, err := store.MarshalKeysShard(data)
	if err != nil {
		return err
	}
	return r.backend.save(store.ShardKeys, raw)
}

func (r *pgKeysRepo) ProviderKeys() []types.ProviderKey {
	data, err := r.read()
	if err != nil {
		return nil
	}
	return store.CloneProviderKeys(data.ProviderKeys)
}

func (r *pgKeysRepo) SetProviderKeys(keys []types.ProviderKey) error {
	data, err := r.read()
	if err != nil {
		return err
	}
	data.ProviderKeys = store.CloneProviderKeys(keys)
	return r.write(data)
}

func (r *pgKeysRepo) PlatformKeys() []types.PlatformKey {
	data, err := r.read()
	if err != nil {
		return nil
	}
	return store.ClonePlatformKeys(data.PlatformKeys)
}

func (r *pgKeysRepo) SetPlatformKeys(keys []types.PlatformKey) error {
	data, err := r.read()
	if err != nil {
		return err
	}
	data.PlatformKeys = store.ClonePlatformKeys(keys)
	return r.write(data)
}

func (r *pgKeysRepo) Approvals() []types.KeyApproval {
	data, err := r.read()
	if err != nil {
		return nil
	}
	return store.CloneApprovals(data.Approvals)
}

func (r *pgKeysRepo) SetApprovals(approvals []types.KeyApproval) error {
	data, err := r.read()
	if err != nil {
		return err
	}
	data.Approvals = store.CloneApprovals(approvals)
	return r.write(data)
}

type pgModelsRepo struct {
	backend shardBackend
}

func (r *pgModelsRepo) read() (store.ModelsShardData, error) {
	raw, err := r.backend.load(store.ShardModels)
	if err != nil {
		return store.ModelsShardData{}, err
	}
	return store.ParseModelsShard(raw)
}

func (r *pgModelsRepo) write(data store.ModelsShardData) error {
	raw, err := store.MarshalModelsShard(data)
	if err != nil {
		return err
	}
	return r.backend.save(store.ShardModels, raw)
}

func (r *pgModelsRepo) Models() []types.ModelInfo {
	data, err := r.read()
	if err != nil {
		return nil
	}
	return store.CloneModels(data.Models)
}

func (r *pgModelsRepo) SetModels(models []types.ModelInfo) error {
	data, err := r.read()
	if err != nil {
		return err
	}
	data.Models = store.CloneModels(models)
	return r.write(data)
}

func (r *pgModelsRepo) RoutingRules() []types.RoutingRule {
	data, err := r.read()
	if err != nil {
		return nil
	}
	return store.CloneRoutingRules(data.RoutingRules)
}

func (r *pgModelsRepo) SetRoutingRules(rules []types.RoutingRule) error {
	data, err := r.read()
	if err != nil {
		return err
	}
	data.RoutingRules = store.CloneRoutingRules(rules)
	return r.write(data)
}

type pgAuditRepo struct {
	backend shardBackend
}

func (r *pgAuditRepo) read() (store.AuditShardData, error) {
	raw, err := r.backend.load(store.ShardAudit)
	if err != nil {
		return store.AuditShardData{}, err
	}
	return store.ParseAuditShard(raw)
}

func (r *pgAuditRepo) write(data store.AuditShardData) error {
	raw, err := store.MarshalAuditShard(data)
	if err != nil {
		return err
	}
	return r.backend.save(store.ShardAudit, raw)
}

func (r *pgAuditRepo) Settings() types.AuditSettings {
	data, err := r.read()
	if err != nil {
		return types.AuditSettings{}
	}
	return data.AuditSettings
}

func (r *pgAuditRepo) SetSettings(settings types.AuditSettings) error {
	data, err := r.read()
	if err != nil {
		return err
	}
	data.AuditSettings = settings
	return r.write(data)
}

func (r *pgAuditRepo) OperationLogs() []types.OperationLog {
	data, err := r.read()
	if err != nil {
		return nil
	}
	return store.CloneOperationLogs(data.OperationLogs)
}

func (r *pgAuditRepo) CallLogs() []types.CallLog {
	data, err := r.read()
	if err != nil {
		return nil
	}
	return store.CloneCallLogs(data.CallLogs)
}

func newDomainRepos(backend shardBackend) (
	store.OrgRepository,
	store.BudgetRepository,
	store.KeysRepository,
	store.ModelsRepository,
	store.AuditRepository,
) {
	return &pgOrgRepo{backend: backend},
		&pgBudgetRepo{backend: backend},
		&pgKeysRepo{backend: backend},
		&pgModelsRepo{backend: backend},
		&pgAuditRepo{backend: backend}
}
