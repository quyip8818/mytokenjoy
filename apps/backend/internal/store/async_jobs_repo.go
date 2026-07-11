package store

import (
	"context"
	"encoding/json"
	"time"
)

const (
	JobStatusPending = "pending"
	JobStatusDone    = "done"
	JobStatusFailed  = "failed"

	JobChannelNewAPISync = "newapi_sync"
	JobChannelRebalance  = "rebalance"
	JobChannelOverrun    = "overrun"
	JobChannelWalletSync = "wallet_sync"

	OutboxKindCreateKey         = "create_key"
	OutboxKindUpdateKey         = "update_key"
	OutboxKindUpsertChannel     = "upsert_channel"
	OutboxKindUpdateModelLimits = "update_model_limits"
	OutboxKindRebalanceKey      = "rebalance_key"

	RebalanceAxisMember      = "member"
	RebalanceAxisDepartment  = "department"
	RebalanceAxisBudgetGroup = "budget_group"
	RebalanceAxisCompany     = "company"
)

const asyncJobClaimLease = 5 * time.Minute

func JobClaimLease() time.Duration {
	return asyncJobClaimLease
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

type WalletSyncQueueEntry struct {
	ID        string
	CompanyID int64
	Status    string
}

type AsyncJobRepository interface {
	EnqueueJob(ctx context.Context, job AsyncJob) error
	ClaimPendingJobs(ctx context.Context, channel string, limit int) ([]AsyncJob, error)
	MarkJobDone(ctx context.Context, id string) error
	MarkJobRetry(ctx context.Context, id string, delay time.Duration, lastError string) error
	MarkJobFailed(ctx context.Context, id string, lastError string) error
}

type NewAPISyncOutboxRepository interface {
	EnqueueNewAPISyncOutbox(ctx context.Context, job AsyncJob) error
	ClaimPendingNewAPISyncOutbox(ctx context.Context, limit int) ([]AsyncJob, error)
}

type OverrunQueueRepository interface {
	EnqueueOverrun(ctx context.Context, payload json.RawMessage) error
	ClaimPendingOverrun(ctx context.Context, limit int) ([]OverrunQueueEntry, error)
	MarkOverrunDone(ctx context.Context, id string) error
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

type AsyncJobsRepository interface {
	AsyncJobRepository
	NewAPISyncOutboxRepository
	RebalanceQueueRepository
	OverrunQueueRepository
	WalletSyncQueueRepository
}
