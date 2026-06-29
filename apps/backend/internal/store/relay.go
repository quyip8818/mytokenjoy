package store

import (
	"encoding/json"
	"time"
)

const (
	RelaySyncStatusPending = "pending"
	RelaySyncStatusSynced  = "synced"
	RelaySyncStatusFailed  = "failed"

	OutboxStatusPending = "pending"
	OutboxStatusDone    = "done"
	OutboxStatusFailed  = "failed"

	OutboxKindCreateToken       = "create_token"
	OutboxKindUpdateToken       = "update_token"
	OutboxKindRevokeToken       = "revoke_token"
	OutboxKindUpsertChannel     = "upsert_channel"
	OutboxKindRebuildAbilities  = "rebuild_abilities"
	OutboxKindUpdateModelLimits = "update_model_limits"
	OutboxKindRebalanceToken    = "rebalance_token"

	RebalanceAxisMember      = "member"
	RebalanceAxisDepartment  = "department"
	RebalanceAxisBudgetGroup = "budget_group"
)

type RelayMapping struct {
	PlatformKeyID    string
	NewAPITokenID    *int64
	MemberID         *string
	DepartmentID     string
	BudgetGroupID    *string
	RelayGroup       string
	SyncStatus       string
	SyncedAt         *time.Time
	RelayRemainQuota *int64
}

type RelayOutboxEntry struct {
	ID        string
	Kind      string
	Payload   json.RawMessage
	Status    string
	Attempts  int
	NextRetry time.Time
	LastError *string
	CreatedAt time.Time
}

type WebhookOutboxEntry struct {
	ID        string
	Payload   json.RawMessage
	Status    string
	Attempts  int
	NextRetry time.Time
	LastError *string
	CreatedAt time.Time
}

type RebalanceQueueEntry struct {
	ID       string
	AxisKind string
	AxisID   string
	Status   string
}

type RelayMappingRepository interface {
	GetMappingByPlatformKeyID(platformKeyID string) (*RelayMapping, error)
	GetMappingByNewAPITokenID(tokenID int64) (*RelayMapping, error)
	ListMappingsByMemberID(memberID string) ([]RelayMapping, error)
	ListMappingsByDepartmentID(departmentID string) ([]RelayMapping, error)
	ListMappingsByBudgetGroupID(budgetGroupID string) ([]RelayMapping, error)
	ListActiveMappings() ([]RelayMapping, error)
	UpsertMapping(mapping RelayMapping) error
	UpdateMappingSync(platformKeyID string, tokenID int64, status string, remainQuota *int64, syncedAt time.Time) error
	UpdateMappingRemainQuota(platformKeyID string, remainQuota int64) error
}

type RelayOutboxRepository interface {
	EnqueueRelayOutbox(entry RelayOutboxEntry) error
	ClaimPendingRelayOutbox(limit int) ([]RelayOutboxEntry, error)
	MarkRelayOutboxDone(id string) error
	MarkRelayOutboxRetry(id string, nextRetry time.Time, lastError string) error
}

type WebhookOutboxRepository interface {
	EnqueueWebhookOutbox(entry WebhookOutboxEntry) error
	ClaimPendingWebhookOutbox(limit int) ([]WebhookOutboxEntry, error)
	MarkWebhookOutboxDone(id string) error
	MarkWebhookOutboxRetry(id string, nextRetry time.Time, lastError string) error
}

type IngestDedupRepository interface {
	HasIngestedLogID(logID int64) (bool, error)
	InsertIngestedLogID(logID int64) error
	GetLastLogID() (int64, error)
	SetLastLogID(logID int64) error
}

type RebalanceQueueRepository interface {
	EnqueueRebalance(axisKind, axisID string) error
	ClaimPendingRebalance(limit int) ([]RebalanceQueueEntry, error)
	MarkRebalanceDone(id string) error
}

type RelayRepository interface {
	RelayMappingRepository
	RelayOutboxRepository
	WebhookOutboxRepository
	IngestDedupRepository
	RebalanceQueueRepository
}
