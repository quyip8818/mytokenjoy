package store

import (
	"encoding/json"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
)

const (
	ShardOrg    = "org"
	ShardBudget = "budget"
	ShardKeys   = "keys"
	ShardModels = "models"
	ShardAudit  = "audit"
)

func AllShardIDs() []string {
	return []string{ShardOrg, ShardBudget, ShardKeys, ShardModels, ShardAudit}
}

type OrgShardData struct {
	DataSourceStatus types.DataSourceStatus `json:"dataSourceStatus"`
	SyncConfig       types.SyncConfig       `json:"syncConfig"`
	SyncLogs         []types.SyncLog        `json:"syncLogs"`
	ImportFailures   []types.ImportFailure  `json:"importFailures"`
	Departments      []types.Department     `json:"departments"`
	Members          []types.Member         `json:"members"`
	Roles            []types.Role           `json:"roles"`
	Permissions      []types.Permission     `json:"permissions"`
}

type BudgetShardData struct {
	BudgetTree       []types.BudgetNode               `json:"budgetTree"`
	BudgetGroups     []types.BudgetGroup              `json:"budgetGroups"`
	OverrunPolicy    types.OverrunPolicyConfig        `json:"overrunPolicy"`
	AlertRules       []types.AlertRule                `json:"alertRules"`
	MemberQuotaPools map[string]types.MemberQuotaPool `json:"memberQuotaPools"`
}

type KeysShardData struct {
	ProviderKeys []types.ProviderKey `json:"providerKeys"`
	PlatformKeys []types.PlatformKey `json:"platformKeys"`
	Approvals    []types.KeyApproval `json:"approvals"`
}

type ModelsShardData struct {
	Models       []types.ModelInfo   `json:"models"`
	RoutingRules []types.RoutingRule `json:"routingRules"`
}

type AuditShardData struct {
	AuditSettings types.AuditSettings  `json:"auditSettings"`
	OperationLogs []types.OperationLog `json:"operationLogs"`
	CallLogs      []types.CallLog      `json:"callLogs"`
}

func ParseOrgShard(raw json.RawMessage) (OrgShardData, error) {
	var data OrgShardData
	if err := json.Unmarshal(raw, &data); err != nil {
		return OrgShardData{}, fmt.Errorf("unmarshal shard %s: %w", ShardOrg, err)
	}
	return data, nil
}

func MarshalOrgShard(data OrgShardData) (json.RawMessage, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal shard %s: %w", ShardOrg, err)
	}
	return raw, nil
}

func ParseBudgetShard(raw json.RawMessage) (BudgetShardData, error) {
	var data BudgetShardData
	if err := json.Unmarshal(raw, &data); err != nil {
		return BudgetShardData{}, fmt.Errorf("unmarshal shard %s: %w", ShardBudget, err)
	}
	return data, nil
}

func MarshalBudgetShard(data BudgetShardData) (json.RawMessage, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal shard %s: %w", ShardBudget, err)
	}
	return raw, nil
}

func ParseKeysShard(raw json.RawMessage) (KeysShardData, error) {
	var data KeysShardData
	if err := json.Unmarshal(raw, &data); err != nil {
		return KeysShardData{}, fmt.Errorf("unmarshal shard %s: %w", ShardKeys, err)
	}
	return data, nil
}

func MarshalKeysShard(data KeysShardData) (json.RawMessage, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal shard %s: %w", ShardKeys, err)
	}
	return raw, nil
}

func ParseModelsShard(raw json.RawMessage) (ModelsShardData, error) {
	var data ModelsShardData
	if err := json.Unmarshal(raw, &data); err != nil {
		return ModelsShardData{}, fmt.Errorf("unmarshal shard %s: %w", ShardModels, err)
	}
	return data, nil
}

func MarshalModelsShard(data ModelsShardData) (json.RawMessage, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal shard %s: %w", ShardModels, err)
	}
	return raw, nil
}

func ParseAuditShard(raw json.RawMessage) (AuditShardData, error) {
	var data AuditShardData
	if err := json.Unmarshal(raw, &data); err != nil {
		return AuditShardData{}, fmt.Errorf("unmarshal shard %s: %w", ShardAudit, err)
	}
	return data, nil
}

func MarshalAuditShard(data AuditShardData) (json.RawMessage, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal shard %s: %w", ShardAudit, err)
	}
	return raw, nil
}

func SnapshotToShards(s Snapshot) (map[string]json.RawMessage, error) {
	payloads := map[string]any{
		ShardOrg: OrgShardData{
			DataSourceStatus: s.DataSourceStatus,
			SyncConfig:       s.SyncConfig,
			SyncLogs:         s.SyncLogs,
			ImportFailures:   s.ImportFailures,
			Departments:      s.Departments,
			Members:          s.Members,
			Roles:            s.Roles,
			Permissions:      s.Permissions,
		},
		ShardBudget: BudgetShardData{
			BudgetTree:       s.BudgetTree,
			BudgetGroups:     s.BudgetGroups,
			OverrunPolicy:    s.OverrunPolicy,
			AlertRules:       s.AlertRules,
			MemberQuotaPools: s.MemberQuotaPools,
		},
		ShardKeys: KeysShardData{
			ProviderKeys: s.ProviderKeys,
			PlatformKeys: s.PlatformKeys,
			Approvals:    s.Approvals,
		},
		ShardModels: ModelsShardData{
			Models:       s.Models,
			RoutingRules: s.RoutingRules,
		},
		ShardAudit: AuditShardData{
			AuditSettings: s.AuditSettings,
			OperationLogs: s.OperationLogs,
			CallLogs:      s.CallLogs,
		},
	}
	shards := make(map[string]json.RawMessage, len(payloads))
	for id, payload := range payloads {
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal shard %s: %w", id, err)
		}
		shards[id] = raw
	}
	return shards, nil
}

func ShardsToSnapshot(shards map[string]json.RawMessage) (Snapshot, error) {
	var snap Snapshot
	for _, id := range AllShardIDs() {
		raw, ok := shards[id]
		if !ok {
			return Snapshot{}, fmt.Errorf("missing shard %s", id)
		}
		if err := mergeShard(id, raw, &snap); err != nil {
			return Snapshot{}, err
		}
	}
	return snap, nil
}

func mergeShard(id string, raw json.RawMessage, snap *Snapshot) error {
	switch id {
	case ShardOrg:
		var payload OrgShardData
		if err := json.Unmarshal(raw, &payload); err != nil {
			return fmt.Errorf("unmarshal shard %s: %w", id, err)
		}
		snap.DataSourceStatus = payload.DataSourceStatus
		snap.SyncConfig = payload.SyncConfig
		snap.SyncLogs = payload.SyncLogs
		snap.ImportFailures = payload.ImportFailures
		snap.Departments = payload.Departments
		snap.Members = payload.Members
		snap.Roles = payload.Roles
		snap.Permissions = payload.Permissions
	case ShardBudget:
		var payload BudgetShardData
		if err := json.Unmarshal(raw, &payload); err != nil {
			return fmt.Errorf("unmarshal shard %s: %w", id, err)
		}
		snap.BudgetTree = payload.BudgetTree
		snap.BudgetGroups = payload.BudgetGroups
		snap.OverrunPolicy = payload.OverrunPolicy
		snap.AlertRules = payload.AlertRules
		snap.MemberQuotaPools = payload.MemberQuotaPools
	case ShardKeys:
		var payload KeysShardData
		if err := json.Unmarshal(raw, &payload); err != nil {
			return fmt.Errorf("unmarshal shard %s: %w", id, err)
		}
		snap.ProviderKeys = payload.ProviderKeys
		snap.PlatformKeys = payload.PlatformKeys
		snap.Approvals = payload.Approvals
	case ShardModels:
		var payload ModelsShardData
		if err := json.Unmarshal(raw, &payload); err != nil {
			return fmt.Errorf("unmarshal shard %s: %w", id, err)
		}
		snap.Models = payload.Models
		snap.RoutingRules = payload.RoutingRules
	case ShardAudit:
		var payload AuditShardData
		if err := json.Unmarshal(raw, &payload); err != nil {
			return fmt.Errorf("unmarshal shard %s: %w", id, err)
		}
		snap.AuditSettings = payload.AuditSettings
		snap.OperationLogs = payload.OperationLogs
		snap.CallLogs = payload.CallLogs
	default:
		return fmt.Errorf("unknown shard %s", id)
	}
	return nil
}

func SnapshotShardFieldsEqual(a, b Snapshot) bool {
	shardsA, err := SnapshotToShards(a)
	if err != nil {
		return false
	}
	shardsB, err := SnapshotToShards(b)
	if err != nil {
		return false
	}
	for _, id := range AllShardIDs() {
		if string(shardsA[id]) != string(shardsB[id]) {
			return false
		}
	}
	return true
}
