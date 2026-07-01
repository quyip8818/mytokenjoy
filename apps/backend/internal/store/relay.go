package store

import (
	"context"
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
	RebalanceAxisCompany     = "company"
)

type RelayMapping struct {
	CompanyID        int64
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
	ID        string
	CompanyID int64
	AxisKind  string
	AxisID    string
	Status    string
}

type RelayMappingRepository interface {
	GetMappingByPlatformKeyID(ctx context.Context, platformKeyID string) (*RelayMapping, error)
	GetMappingByFullKey(ctx context.Context, fullKey string) (*RelayMapping, error)
	GetMappingByNewAPITokenID(ctx context.Context, tokenID int64) (*RelayMapping, error)
	FindMappingByNewAPITokenID(ctx context.Context, tokenID int64) (*RelayMapping, error)
	ListMappingsByMemberID(ctx context.Context, memberID string) ([]RelayMapping, error)
	ListMappingsByDepartmentID(ctx context.Context, departmentID string) ([]RelayMapping, error)
	ListMappingsByBudgetGroupID(ctx context.Context, budgetGroupID string) ([]RelayMapping, error)
	ListActiveMappings(ctx context.Context) ([]RelayMapping, error)
	ListActiveMappingsByCompany(ctx context.Context, companyID int64) ([]RelayMapping, error)
	UpsertMapping(ctx context.Context, mapping RelayMapping) error
	UpdateMappingSync(ctx context.Context, platformKeyID string, tokenID int64, status string, remainQuota *int64, syncedAt time.Time) error
	UpdateMappingRemainQuota(ctx context.Context, platformKeyID string, remainQuota int64) error
}

type RelayOutboxRepository interface {
	EnqueueRelayOutbox(ctx context.Context, entry RelayOutboxEntry) error
	ClaimPendingRelayOutbox(ctx context.Context, limit int) ([]RelayOutboxEntry, error)
	MarkRelayOutboxDone(ctx context.Context, id string) error
	MarkRelayOutboxRetry(ctx context.Context, id string, nextRetry time.Time, lastError string) error
}

type WebhookOutboxRepository interface {
	EnqueueWebhookOutbox(ctx context.Context, entry WebhookOutboxEntry) error
	ClaimPendingWebhookOutbox(ctx context.Context, limit int) ([]WebhookOutboxEntry, error)
	MarkWebhookOutboxDone(ctx context.Context, id string) error
	MarkWebhookOutboxRetry(ctx context.Context, id string, nextRetry time.Time, lastError string) error
}

type IngestDedupRepository interface {
	HasIngestedLogID(ctx context.Context, logID int64) (bool, error)
	InsertIngestedLogID(ctx context.Context, logID int64) error
	GetLastLogID(ctx context.Context) (int64, error)
	SetLastLogID(ctx context.Context, logID int64) error
}

type RebalanceQueueRepository interface {
	EnqueueRebalance(ctx context.Context, axisKind, axisID string) error
	ClaimPendingRebalance(ctx context.Context, limit int) ([]RebalanceQueueEntry, error)
	MarkRebalanceDone(ctx context.Context, id string) error
}

type RelayRepository interface {
	RelayMappingRepository
	RelayOutboxRepository
	WebhookOutboxRepository
	IngestDedupRepository
	RebalanceQueueRepository
}
