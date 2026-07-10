package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

const (
	RelaySyncStatusPending = "pending"
	RelaySyncStatusSynced  = "synced"
	RelaySyncStatusFailed  = "failed"

	JobStatusPending = "pending"
	JobStatusDone    = "done"
	JobStatusFailed  = "failed"

	JobChannelRelay      = "relay"
	JobChannelRebalance  = "rebalance"
	JobChannelOverrun    = "overrun"
	JobChannelWalletSync = "wallet_sync"

	OutboxStatusPending = JobStatusPending
	OutboxStatusDone    = JobStatusDone
	OutboxStatusFailed  = JobStatusFailed

	OutboxChannelRelay = JobChannelRelay

	OutboxKindCreateToken       = "create_token"
	OutboxKindUpdateToken       = "update_token"
	OutboxKindUpsertChannel     = "upsert_channel"
	OutboxKindUpdateModelLimits = "update_model_limits"
	OutboxKindRebalanceToken    = "rebalance_token"

	RebalanceAxisMember      = "member"
	RebalanceAxisDepartment  = "department"
	RebalanceAxisBudgetGroup = "budget_group"
	RebalanceAxisCompany     = "company"
)

const asyncJobClaimLease = 5 * time.Minute

func JobClaimLease() time.Duration {
	return asyncJobClaimLease
}

func HashPlatformKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

type RelayMapping struct {
	CompanyID              int64
	PlatformKeyID          string
	NewAPITokenID          *int64
	MemberID               *string
	DepartmentID           string
	BudgetGroupID          *string
	RelayGroup             string
	SyncStatus             string
	SyncedAt               *time.Time
	NewAPITokenRemainQuota *int64
}

type AsyncJob struct {
	ID        string
	CompanyID *int64
	Channel   string
	Kind      string
	DedupeKey *string
	Payload   json.RawMessage
	Status    string
	Attempts  int
	NextRetry time.Time
	LastError *string
	CreatedAt time.Time
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

type RebalanceQueueEntry struct {
	ID        string
	CompanyID int64
	AxisKind  string
	AxisID    string
	Status    string
}

type OverrunQueueEntry struct {
	ID        string
	CompanyID int64
	Payload   json.RawMessage
	Status    string
}

type RelayMappingRepository interface {
	GetMappingByPlatformKeyID(ctx context.Context, platformKeyID string) (*RelayMapping, error)
	GetMappingByKeyHash(ctx context.Context, keyHash string) (*RelayMapping, error)
	GetMappingByNewAPITokenID(ctx context.Context, tokenID int64) (*RelayMapping, error)
	FindMappingByNewAPITokenID(ctx context.Context, tokenID int64) (*RelayMapping, error)
	ListMappingsByMemberID(ctx context.Context, memberID string) ([]RelayMapping, error)
	ListMappingsByDepartmentID(ctx context.Context, departmentID string) ([]RelayMapping, error)
	ListMappingsByBudgetGroupID(ctx context.Context, budgetGroupID string) ([]RelayMapping, error)
	ListActiveMappings(ctx context.Context) ([]RelayMapping, error)
	ListActiveMappingsByCompany(ctx context.Context, companyID int64) ([]RelayMapping, error)
	UpsertMapping(ctx context.Context, mapping RelayMapping) error
	UpdateMappingSync(ctx context.Context, platformKeyID string, tokenID int64, status string, remainQuota *int64, syncedAt time.Time) error
	UpdateMappingNewAPITokenRemainQuota(ctx context.Context, platformKeyID string, remainQuota int64) error
}

type AsyncJobRepository interface {
	EnqueueJob(ctx context.Context, job AsyncJob) error
	ClaimPendingJobs(ctx context.Context, channel string, limit int) ([]AsyncJob, error)
	MarkJobDone(ctx context.Context, id string) error
	MarkJobRetry(ctx context.Context, id string, delay time.Duration, lastError string) error
}

type RelayOutboxRepository interface {
	EnqueueRelayOutbox(ctx context.Context, entry RelayOutboxEntry) error
	ClaimPendingRelayOutbox(ctx context.Context, limit int) ([]RelayOutboxEntry, error)
	MarkRelayOutboxDone(ctx context.Context, id string) error
	MarkRelayOutboxRetry(ctx context.Context, id string, delay time.Duration, lastError string) error
	MarkRelayOutboxFailed(ctx context.Context, id string, lastError string) error
}

type OverrunQueueRepository interface {
	EnqueueOverrun(ctx context.Context, payload json.RawMessage) error
	ClaimPendingOverrun(ctx context.Context, limit int) ([]OverrunQueueEntry, error)
	MarkOverrunDone(ctx context.Context, id string) error
}

type WalletSyncQueueEntry struct {
	ID        string
	CompanyID int64
	Status    string
}

type WalletSyncQueueRepository interface {
	EnqueueWalletSync(ctx context.Context, companyID int64) error
	ClaimPendingWalletSync(ctx context.Context, limit int) ([]WalletSyncQueueEntry, error)
	MarkWalletSyncDone(ctx context.Context, id string) error
	HasPendingWalletSync(ctx context.Context, companyID int64) (bool, error)
}

type RebalanceQueueRepository interface {
	EnqueueRebalance(ctx context.Context, axisKind, axisID string) error
	ClaimPendingRebalance(ctx context.Context, limit int) ([]RebalanceQueueEntry, error)
	MarkRebalanceDone(ctx context.Context, id string) error
}

type RelayJobRepository interface {
	RelayOutboxRepository
	RebalanceQueueRepository
	OverrunQueueRepository
	WalletSyncQueueRepository
}

type RelayRepository interface {
	RelayMappingRepository
	RelayJobRepository
	AsyncJobRepository
}
