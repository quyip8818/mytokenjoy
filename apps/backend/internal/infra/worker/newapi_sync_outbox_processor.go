package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *Runner) workerCtx(ctx context.Context, companyID int64) context.Context {
	return company.WithDefaultCompany(ctx, companyID)
}

func (r *Runner) processNewAPISyncOutbox(ctx context.Context) error {
	workerCtx := r.workerCtx(ctx, r.cfg.LocalCompanyID)
	jobs, err := r.asyncJobs.ClaimPendingNewAPISyncOutbox(workerCtx, 20)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		var processErr error
		switch job.Kind {
		case store.OutboxKindCreateKey:
			var payload newapisync.CreateKeyOutboxPayload
			if err := json.Unmarshal(job.Payload, &payload); err != nil {
				processErr = fmt.Errorf("unmarshal create key payload: %w", err)
				break
			}
			entryCtx := r.workerCtx(ctx, payload.CompanyID)
			_, processErr = r.newAPISync.TrySyncCreate(entryCtx, payload.PlatformKeyID)
		case store.OutboxKindUpdateKey:
			var payload newapisync.UpdateKeyOutboxPayload
			if err := json.Unmarshal(job.Payload, &payload); err != nil {
				processErr = fmt.Errorf("unmarshal update key payload: %w", err)
				break
			}
			entryCtx := r.workerCtx(ctx, payload.CompanyID)
			processErr = r.newAPISync.SyncUpdatePlatformKey(entryCtx, payload.PlatformKeyID, nil)
		case store.OutboxKindUpsertChannel:
			var payload newapisync.UpsertChannelOutboxPayload
			if err := json.Unmarshal(job.Payload, &payload); err != nil {
				processErr = fmt.Errorf("unmarshal upsert channel payload: %w", err)
				break
			}
			entryCtx := r.workerCtx(ctx, payload.CompanyID)
			processErr = r.newAPISync.SyncUpsertProviderKey(entryCtx, payload.ProviderKeyID)
		case store.OutboxKindUpdateModelLimits:
			var payload newapisync.UpdateModelLimitsOutboxPayload
			if err := json.Unmarshal(job.Payload, &payload); err != nil {
				processErr = fmt.Errorf("unmarshal update model limits payload: %w", err)
				break
			}
			entryCtx := r.workerCtx(ctx, payload.CompanyID)
			processErr = r.newAPISync.SyncModelLimitsForDepartment(entryCtx, payload.DepartmentID)
		case store.OutboxKindRebalanceKey:
			var payload newapisync.RebalanceAxisOutboxPayload
			if err := json.Unmarshal(job.Payload, &payload); err != nil {
				processErr = fmt.Errorf("unmarshal rebalance payload: %w", err)
				break
			}
			entryCtx := r.workerCtx(ctx, payload.CompanyID)
			processErr = r.rebalance.ProcessAxis(entryCtx, payload.AxisKind, payload.AxisID)
		default:
			processErr = fmt.Errorf("unknown newapi sync outbox kind: %s", job.Kind)
		}
		if processErr != nil {
			r.finishNewAPISyncOutboxJob(workerCtx, job, processErr)
			continue
		}
		r.markJobDone(workerCtx, job.ID)
	}
	return nil
}

func (r *Runner) finishNewAPISyncOutboxJob(ctx context.Context, job store.AsyncJob, processErr error) {
	if newapisync.IsPermanentOutboxError(processErr) {
		r.markJobFailed(ctx, job.ID, processErr.Error())
		return
	}
	r.markJobRetry(ctx, job.ID, backoff(job.Attempts), processErr.Error())
}
