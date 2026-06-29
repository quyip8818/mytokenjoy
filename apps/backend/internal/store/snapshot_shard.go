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

type orgShardPayload struct {
	DataSourceStatus types.DataSourceStatus `json:"dataSourceStatus"`
	SyncConfig       types.SyncConfig       `json:"syncConfig"`
	SyncLogs         []types.SyncLog        `json:"syncLogs"`
	ImportFailures   []types.ImportFailure  `json:"importFailures"`
	Departments      []types.Department     `json:"departments"`
	Members          []types.Member         `json:"members"`
	Roles            []types.Role           `json:"roles"`
	Permissions      []types.Permission     `json:"permissions"`
}

type budgetShardPayload struct {
	BudgetTree       []types.BudgetNode               `json:"budgetTree"`
	BudgetGroups     []types.BudgetGroup              `json:"budgetGroups"`
	OverrunPolicy    types.OverrunPolicyConfig        `json:"overrunPolicy"`
	AlertRules       []types.AlertRule                `json:"alertRules"`
	MemberQuotaPools map[string]types.MemberQuotaPool `json:"memberQuotaPools"`
}

type keysShardPayload struct {
	ProviderKeys []types.ProviderKey `json:"providerKeys"`
	PlatformKeys []types.PlatformKey `json:"platformKeys"`
	Approvals    []types.KeyApproval `json:"approvals"`
}

type modelsShardPayload struct {
	Models       []types.ModelInfo   `json:"models"`
	RoutingRules []types.RoutingRule `json:"routingRules"`
}

type auditShardPayload struct {
	AuditSettings types.AuditSettings  `json:"auditSettings"`
	OperationLogs []types.OperationLog `json:"operationLogs"`
	CallLogs      []types.CallLog      `json:"callLogs"`
}

func SnapshotToShards(s Snapshot) (map[string]json.RawMessage, error) {
	payloads := map[string]any{
		ShardOrg: orgShardPayload{
			DataSourceStatus: s.DataSourceStatus,
			SyncConfig:       s.SyncConfig,
			SyncLogs:         s.SyncLogs,
			ImportFailures:   s.ImportFailures,
			Departments:      s.Departments,
			Members:          s.Members,
			Roles:            s.Roles,
			Permissions:      s.Permissions,
		},
		ShardBudget: budgetShardPayload{
			BudgetTree:       s.BudgetTree,
			BudgetGroups:     s.BudgetGroups,
			OverrunPolicy:    s.OverrunPolicy,
			AlertRules:       s.AlertRules,
			MemberQuotaPools: s.MemberQuotaPools,
		},
		ShardKeys: keysShardPayload{
			ProviderKeys: s.ProviderKeys,
			PlatformKeys: s.PlatformKeys,
			Approvals:    s.Approvals,
		},
		ShardModels: modelsShardPayload{
			Models:       s.Models,
			RoutingRules: s.RoutingRules,
		},
		ShardAudit: auditShardPayload{
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
		var payload orgShardPayload
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
		var payload budgetShardPayload
		if err := json.Unmarshal(raw, &payload); err != nil {
			return fmt.Errorf("unmarshal shard %s: %w", id, err)
		}
		snap.BudgetTree = payload.BudgetTree
		snap.BudgetGroups = payload.BudgetGroups
		snap.OverrunPolicy = payload.OverrunPolicy
		snap.AlertRules = payload.AlertRules
		snap.MemberQuotaPools = payload.MemberQuotaPools
	case ShardKeys:
		var payload keysShardPayload
		if err := json.Unmarshal(raw, &payload); err != nil {
			return fmt.Errorf("unmarshal shard %s: %w", id, err)
		}
		snap.ProviderKeys = payload.ProviderKeys
		snap.PlatformKeys = payload.PlatformKeys
		snap.Approvals = payload.Approvals
	case ShardModels:
		var payload modelsShardPayload
		if err := json.Unmarshal(raw, &payload); err != nil {
			return fmt.Errorf("unmarshal shard %s: %w", id, err)
		}
		snap.Models = payload.Models
		snap.RoutingRules = payload.RoutingRules
	case ShardAudit:
		var payload auditShardPayload
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
