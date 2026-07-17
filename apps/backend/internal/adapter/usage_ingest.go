package adapter

import (
	"context"

	"github.com/google/uuid"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/store"
)

type usageIngestEnqueuer struct {
	enqueuer jobs.Enqueuer
}

// NewUsageIngestEnqueuer adapts infra/jobs.Enqueuer to domain/usage.IngestJobEnqueuer.
func NewUsageIngestEnqueuer(enqueuer jobs.Enqueuer) domainusage.IngestJobEnqueuer {
	return usageIngestEnqueuer{enqueuer: JobsOrNoop(enqueuer)}
}

func (u usageIngestEnqueuer) EnqueueAfterIngest(ctx context.Context, tx store.Tx, companyID uuid.UUID, effects *domainusage.IngestEffects) error {
	// Dashboard projection is now driven purely by the periodic watchdog (hourly).
	// No per-ingest enqueue needed — usage_buckets are hour-granularity anyway.
	if err := jobs.InsertWalletSync(ctx, u.enqueuer, tx, companyID); err != nil {
		return err
	}
	// Conditional overrun: only enqueue when remain is known <= 0 or unknown.
	if effects != nil && domainusage.ShouldEnqueueOverrun(effects.Summaries, platformKeyIDFromEffects(effects)) {
		payload := domainusage.OverrunPayloadFromEffects(effects)
		if payload != nil {
			if err := jobs.InsertOverrun(ctx, u.enqueuer, tx, companyID, payload); err != nil {
				return err
			}
		}
	}
	return nil
}

func platformKeyIDFromEffects(effects *domainusage.IngestEffects) uuid.UUID {
	if effects == nil {
		return uuid.Nil
	}
	return effects.PlatformKeyID
}

var _ domainusage.IngestJobEnqueuer = usageIngestEnqueuer{}
