package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *Runner) workerCtx(ctx context.Context, companyID int64) context.Context {
	return company.WithDefaultCompany(ctx, companyID)
}

func (r *Runner) processRelayOutbox(ctx context.Context) error {
	workerCtx := r.workerCtx(ctx, r.cfg.DefaultCompanyID)
	entries, err := r.relayOutbox.ClaimPendingRelayOutbox(workerCtx, 20)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		var processErr error
		switch entry.Kind {
		case store.OutboxKindCreateToken:
			var payload relay.CreateTokenOutboxPayload
			if err := json.Unmarshal(entry.Payload, &payload); err != nil {
				processErr = fmt.Errorf("unmarshal create token payload: %w", err)
				break
			}
			entryCtx := r.workerCtx(ctx, payload.CompanyID)
			_, processErr = r.relaySync.TrySyncCreate(entryCtx, payload.PlatformKeyID)
		case store.OutboxKindUpdateToken:
			var payload relay.UpdateTokenOutboxPayload
			if err := json.Unmarshal(entry.Payload, &payload); err != nil {
				processErr = fmt.Errorf("unmarshal update token payload: %w", err)
				break
			}
			entryCtx := r.workerCtx(ctx, payload.CompanyID)
			processErr = r.relaySync.SyncUpdatePlatformKey(entryCtx, payload.PlatformKeyID)
		case store.OutboxKindRevokeToken:
			var payload relay.UpdateTokenOutboxPayload
			if err := json.Unmarshal(entry.Payload, &payload); err != nil {
				processErr = fmt.Errorf("unmarshal revoke token payload: %w", err)
				break
			}
			entryCtx := r.workerCtx(ctx, payload.CompanyID)
			processErr = r.relaySync.SyncRevokePlatformKey(entryCtx, payload.PlatformKeyID)
		case store.OutboxKindUpsertChannel:
			var payload relay.UpsertChannelOutboxPayload
			if err := json.Unmarshal(entry.Payload, &payload); err != nil {
				processErr = fmt.Errorf("unmarshal upsert channel payload: %w", err)
				break
			}
			entryCtx := r.workerCtx(ctx, payload.CompanyID)
			processErr = r.relaySync.SyncUpsertProviderKey(entryCtx, payload.ProviderKeyID)
		case store.OutboxKindUpdateModelLimits:
			var payload relay.UpdateModelLimitsOutboxPayload
			if err := json.Unmarshal(entry.Payload, &payload); err != nil {
				processErr = fmt.Errorf("unmarshal update model limits payload: %w", err)
				break
			}
			entryCtx := r.workerCtx(ctx, payload.CompanyID)
			processErr = r.relaySync.SyncModelLimitsForDepartment(entryCtx, payload.DepartmentID)
		case store.OutboxKindRebalanceToken:
			var payload relay.RebalanceAxisOutboxPayload
			if err := json.Unmarshal(entry.Payload, &payload); err != nil {
				processErr = fmt.Errorf("unmarshal rebalance payload: %w", err)
				break
			}
			entryCtx := r.workerCtx(ctx, payload.CompanyID)
			processErr = r.rebalance.ProcessAxis(entryCtx, payload.AxisKind, payload.AxisID)
		default:
			processErr = fmt.Errorf("unknown relay outbox kind: %s", entry.Kind)
		}
		if processErr != nil {
			next := time.Now().Add(backoff(entry.Attempts))
			r.markRelayOutboxRetry(workerCtx, entry.ID, next, processErr.Error())
			continue
		}
		r.markRelayOutboxDone(workerCtx, entry.ID)
	}
	return nil
}
