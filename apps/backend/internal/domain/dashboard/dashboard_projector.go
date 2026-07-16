package dashboard

import (
	"context"
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

const defaultDashboardBatchSize = 500

// ProjectorStore is the narrow store surface the dashboard projector needs.
type ProjectorStore interface {
	Company() store.CompanyRepository
	WithTx(ctx context.Context, fn func(store.Store) error) error
}

type Projector struct {
	cfg       config.Config
	store     ProjectorStore
	enqueuer  JobEnqueuer
	batchSize int
	logger    *slog.Logger
}

func NewProjector(cfg config.Config, st ProjectorStore, enqueuer JobEnqueuer, logger *slog.Logger) *Projector {
	if enqueuer == nil {
		enqueuer = NoopJobEnqueuer
	}
	return &Projector{
		cfg:       cfg,
		store:     st,
		enqueuer:  enqueuer,
		batchSize: defaultDashboardBatchSize,
		logger:    logger,
	}
}

func (p *Projector) RunBatch(ctx context.Context, companyID int64) (bool, error) {
	co, err := p.store.Company().GetByID(ctx, companyID)
	if err != nil {
		return false, err
	}
	if co == nil {
		return false, nil
	}
	ctx = company.WithContext(ctx, company.ContextFromStore(*co))

	var count int
	err = p.store.WithTx(ctx, func(tx store.Store) error {
		progress, err := tx.DashboardProjectionProgress().GetForUpdate(ctx, store.DashboardProjectionStream)
		if err != nil {
			return err
		}
		cursor := store.LedgerProjectorCursor{
			LastOccurredAt: progress.LastOccurredAt,
			LastLedgerID:   progress.LastLedgerID,
			Limit:          p.batchSize,
		}
		batch, err := tx.Ledger().ListCallSettledAfterCursor(ctx, cursor)
		if err != nil {
			return err
		}
		if len(batch) == 0 {
			return nil
		}
		for _, entry := range batch {
			if err := upsertDashboardBucket(ctx, tx.Usage(), entry); err != nil {
				return err
			}
		}
		last := batch[len(batch)-1]
		if err := tx.DashboardProjectionProgress().Advance(ctx, store.DashboardProjectionStream, last.OccurredAt, last.ID); err != nil {
			return err
		}
		count = len(batch)
		return nil
	})
	if err != nil {
		return false, err
	}
	if count == 0 {
		return false, nil
	}
	hasMore := count >= p.batchSize
	if hasMore {
		if err := p.enqueuer.InsertDashboardProject(ctx, companyID); err != nil {
			return false, err
		}
	}
	return hasMore, nil
}

func upsertDashboardBucket(ctx context.Context, usage store.UsageRepository, entry types.UsageLedgerEntry) error {
	return usage.UpsertBucket(ctx, bucketFromLedgerEntry(entry))
}
