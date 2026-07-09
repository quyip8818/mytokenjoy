package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *relayRepo) EnqueueJob(ctx context.Context, job store.AsyncJob) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO async_jobs (id, company_id, channel, kind, dedupe_key, payload, status, attempts, next_retry, last_error, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
		ON CONFLICT (company_id, channel, dedupe_key)
		WHERE dedupe_key IS NOT NULL AND status = 'pending'
		DO NOTHING
	`, job.ID, job.CompanyID, job.Channel, job.Kind, job.DedupeKey, job.Payload, job.Status, job.Attempts, job.NextRetry, job.LastError, job.CreatedAt)
	return err
}

func (r *relayRepo) ClaimPendingJobs(ctx context.Context, channel string, limit int) ([]store.AsyncJob, error) {
	leaseUntil := time.Now().Add(store.JobClaimLease())
	rows, err := r.db.Query(ctx, `
		WITH claimed AS (
			SELECT id
			FROM async_jobs
			WHERE channel = $1 AND status = $2 AND next_retry <= NOW()
			ORDER BY created_at
			LIMIT $3
			FOR UPDATE SKIP LOCKED
		)
		UPDATE async_jobs AS j
		SET next_retry = $4, updated_at = NOW()
		FROM claimed
		WHERE j.id = claimed.id
		RETURNING j.id, j.company_id, j.channel, j.kind, j.dedupe_key, j.payload, j.status, j.attempts, j.next_retry, j.last_error, j.created_at
	`, channel, store.JobStatusPending, limit, leaseUntil)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAsyncJobs(rows)
}

func scanAsyncJobs(rows pgx.Rows) ([]store.AsyncJob, error) {
	out := make([]store.AsyncJob, 0)
	for rows.Next() {
		var j store.AsyncJob
		if err := rows.Scan(&j.ID, &j.CompanyID, &j.Channel, &j.Kind, &j.DedupeKey, &j.Payload, &j.Status, &j.Attempts, &j.NextRetry, &j.LastError, &j.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

func (r *relayRepo) MarkJobDone(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE async_jobs SET status = $2, updated_at = NOW() WHERE id = $1
	`, id, store.JobStatusDone)
	return err
}

func (r *relayRepo) MarkJobRetry(ctx context.Context, id string, nextRetry time.Time, lastError string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE async_jobs SET attempts = attempts + 1, next_retry = $2, last_error = $3, updated_at = NOW()
		WHERE id = $1
	`, id, nextRetry, lastError)
	return err
}

func (r *relayRepo) EnqueueRelayOutbox(ctx context.Context, entry store.RelayOutboxEntry) error {
	return r.EnqueueJob(ctx, store.AsyncJob{
		ID:        entry.ID,
		Channel:   store.JobChannelRelay,
		Kind:      entry.Kind,
		Payload:   entry.Payload,
		Status:    entry.Status,
		Attempts:  entry.Attempts,
		NextRetry: entry.NextRetry,
		LastError: entry.LastError,
		CreatedAt: entry.CreatedAt,
	})
}

func (r *relayRepo) ClaimPendingRelayOutbox(ctx context.Context, limit int) ([]store.RelayOutboxEntry, error) {
	jobs, err := r.ClaimPendingJobs(ctx, store.JobChannelRelay, limit)
	if err != nil {
		return nil, err
	}
	out := make([]store.RelayOutboxEntry, len(jobs))
	for i, j := range jobs {
		out[i] = store.RelayOutboxEntry{
			ID:        j.ID,
			Kind:      j.Kind,
			Payload:   j.Payload,
			Status:    j.Status,
			Attempts:  j.Attempts,
			NextRetry: j.NextRetry,
			LastError: j.LastError,
			CreatedAt: j.CreatedAt,
		}
	}
	return out, nil
}

func (r *relayRepo) MarkRelayOutboxDone(ctx context.Context, id string) error {
	return r.MarkJobDone(ctx, id)
}

func (r *relayRepo) MarkRelayOutboxRetry(ctx context.Context, id string, nextRetry time.Time, lastError string) error {
	return r.MarkJobRetry(ctx, id, nextRetry, lastError)
}

func (r *relayRepo) MarkRelayOutboxFailed(ctx context.Context, id string, lastError string) error {
	return r.MarkJobFailed(ctx, id, lastError)
}

func (r *relayRepo) MarkJobFailed(ctx context.Context, id string, lastError string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE async_jobs SET status = $2, attempts = attempts + 1, last_error = $3, updated_at = NOW()
		WHERE id = $1
	`, id, store.JobStatusFailed, lastError)
	return err
}

func (r *relayRepo) EnqueueRebalance(ctx context.Context, axisKind, axisID string) error {
	companyID := store.CompanyID(ctx)
	dedupe := fmt.Sprintf("%s:%s", axisKind, axisID)
	id := fmt.Sprintf("rb-%d-%s-%d", companyID, dedupe, time.Now().UnixNano())
	return r.EnqueueJob(ctx, store.AsyncJob{
		ID:        id,
		CompanyID: &companyID,
		Channel:   store.JobChannelRebalance,
		Kind:      store.OutboxKindRebalanceToken,
		DedupeKey: &dedupe,
		Payload:   json.RawMessage(fmt.Sprintf(`{"axis_kind":%q,"axis_id":%q}`, axisKind, axisID)),
		Status:    store.JobStatusPending,
		CreatedAt: time.Now().UTC(),
		NextRetry: time.Now().UTC(),
	})
}

func (r *relayRepo) ClaimPendingRebalance(ctx context.Context, limit int) ([]store.RebalanceQueueEntry, error) {
	jobs, err := r.ClaimPendingJobs(ctx, store.JobChannelRebalance, limit)
	if err != nil {
		return nil, err
	}
	out := make([]store.RebalanceQueueEntry, 0, len(jobs))
	for _, j := range jobs {
		if j.CompanyID == nil {
			continue
		}
		var payload struct {
			AxisKind string `json:"axis_kind"`
			AxisID   string `json:"axis_id"`
		}
		_ = json.Unmarshal(j.Payload, &payload)
		axisKind := payload.AxisKind
		axisID := payload.AxisID
		if axisKind == "" {
			axisKind = j.Kind
		}
		out = append(out, store.RebalanceQueueEntry{
			ID:        j.ID,
			CompanyID: *j.CompanyID,
			AxisKind:  axisKind,
			AxisID:    axisID,
			Status:    j.Status,
		})
	}
	return out, nil
}

func (r *relayRepo) MarkRebalanceDone(ctx context.Context, id string) error {
	return r.MarkJobDone(ctx, id)
}

func (r *relayRepo) EnqueueOverrun(ctx context.Context, payload json.RawMessage) error {
	companyID := store.CompanyID(ctx)
	id := fmt.Sprintf("ovr-%d-%d", companyID, time.Now().UnixNano())
	return r.EnqueueJob(ctx, store.AsyncJob{
		ID:        id,
		CompanyID: &companyID,
		Channel:   store.JobChannelOverrun,
		Kind:      "overrun",
		Payload:   payload,
		Status:    store.JobStatusPending,
		CreatedAt: time.Now().UTC(),
		NextRetry: time.Now().UTC(),
	})
}

func (r *relayRepo) ClaimPendingOverrun(ctx context.Context, limit int) ([]store.OverrunQueueEntry, error) {
	jobs, err := r.ClaimPendingJobs(ctx, store.JobChannelOverrun, limit)
	if err != nil {
		return nil, err
	}
	out := make([]store.OverrunQueueEntry, 0, len(jobs))
	for _, j := range jobs {
		if j.CompanyID == nil {
			continue
		}
		out = append(out, store.OverrunQueueEntry{
			ID:        j.ID,
			CompanyID: *j.CompanyID,
			Payload:   j.Payload,
			Status:    j.Status,
		})
	}
	return out, nil
}

func (r *relayRepo) MarkOverrunDone(ctx context.Context, id string) error {
	return r.MarkJobDone(ctx, id)
}

func walletSyncDedupeKey(companyID int64) string {
	return fmt.Sprintf("wallet_sync:%d", companyID)
}

func (r *relayRepo) EnqueueWalletSync(ctx context.Context, companyID int64) error {
	dedupe := walletSyncDedupeKey(companyID)
	id := fmt.Sprintf("ws-%d-%d", companyID, time.Now().UnixNano())
	payload, err := json.Marshal(map[string]int64{"company_id": companyID})
	if err != nil {
		return err
	}
	debounceUntil := time.Now().UTC().Add(common.WalletSyncDebounceSecs * time.Second)
	_, err = r.db.Exec(ctx, `
		INSERT INTO async_jobs (id, company_id, channel, kind, dedupe_key, payload, status, attempts, next_retry, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 0, $8, NOW(), NOW())
		ON CONFLICT (company_id, channel, dedupe_key)
		WHERE dedupe_key IS NOT NULL AND status = 'pending'
		DO UPDATE SET next_retry = GREATEST(async_jobs.next_retry, EXCLUDED.next_retry), updated_at = NOW()
	`, id, companyID, store.JobChannelWalletSync, "wallet_sync", dedupe, payload, store.JobStatusPending, debounceUntil)
	return err
}

func (r *relayRepo) ClaimPendingWalletSync(ctx context.Context, limit int) ([]store.WalletSyncQueueEntry, error) {
	jobs, err := r.ClaimPendingJobs(ctx, store.JobChannelWalletSync, limit)
	if err != nil {
		return nil, err
	}
	out := make([]store.WalletSyncQueueEntry, 0, len(jobs))
	for _, j := range jobs {
		if j.CompanyID == nil {
			continue
		}
		out = append(out, store.WalletSyncQueueEntry{
			ID:        j.ID,
			CompanyID: *j.CompanyID,
			Status:    j.Status,
		})
	}
	return out, nil
}

func (r *relayRepo) MarkWalletSyncDone(ctx context.Context, id string) error {
	return r.MarkJobDone(ctx, id)
}

func (r *relayRepo) HasPendingWalletSync(ctx context.Context, companyID int64) (bool, error) {
	dedupe := walletSyncDedupeKey(companyID)
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM async_jobs
			WHERE channel = $1 AND dedupe_key = $2 AND status = $3
		)
	`, store.JobChannelWalletSync, dedupe, store.JobStatusPending).Scan(&exists)
	return exists, err
}
