package dashboard

import (
	"context"
	"log/slog"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

const defaultDashboardBatchSize = 500

type Projector struct {
	cfg       config.Config
	store     store.Store
	enqueuer  JobEnqueuer
	batchSize int
	logger    *slog.Logger
}

func NewProjector(cfg config.Config, st store.Store, enqueuer JobEnqueuer, logger *slog.Logger) *Projector {
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
	ctx = company.WithContext(ctx, companyContextFromStore(*co))

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
	var memberID string
	if entry.MemberID != nil {
		memberID = *entry.MemberID
	}
	return usage.UpsertBucket(ctx, types.UsageBucketRow{
		BucketStart:  entry.OccurredAt.UTC().Truncate(time.Hour),
		DepartmentID: entry.DepartmentID,
		MemberID:     memberID,
		Model:        entry.Model,
		Cost:         entry.Amount,
		CallCount:    1,
		InputTokens:  entry.InputTokens,
		OutputTokens: entry.OutputTokens,
	})
}

func companyContextFromStore(co store.Company) company.Context {
	info := company.Context{
		CompanyID: co.ID,
		Slug:      co.Slug,
		Status:    co.Status,
	}
	if co.NewAPIWalletUserID != nil {
		info.NewAPIWalletUserID = *co.NewAPIWalletUserID
	}
	return info
}
