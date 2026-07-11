package budget

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/infra/budgetcheck"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

const defaultProjectorBatchSize = 500

type Projector struct {
	cfg          config.Config
	store        store.Store
	enqueuer     jobs.Enqueuer
	batchSize    int
	logger       *slog.Logger
	gatewayCache budgetcheck.Store
}

func (p *Projector) RunBatch(ctx context.Context, companyID int64) (bool, error) {
	co, err := p.store.Company().GetByID(ctx, companyID)
	if err != nil {
		return false, err
	}
	if co == nil {
		return false, nil
	}
	ctx = company.WithContext(ctx, companyFromStore(*co))

	var entries []types.UsageLedgerEntry
	err = p.store.WithTx(ctx, func(tx store.Store) error {
		progress, err := tx.BudgetProjectionProgress().GetForUpdate(ctx, store.BudgetProjectionStream)
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
		nodes := tx.Org().Nodes()
		for _, entry := range batch {
			open, err := pkgbudget.OpenDepartmentPeriod(ctx, nodes, entry.DepartmentID, p.cfg.Clock())
			if err != nil {
				return err
			}
			if err := ApplyIncrement(ctx, tx.BudgetConsumed(), nodes, entry, open); err != nil {
				return err
			}
		}
		last := batch[len(batch)-1]
		if err := tx.BudgetProjectionProgress().Advance(ctx, store.BudgetProjectionStream, last.OccurredAt, last.ID); err != nil {
			return err
		}
		entries = batch
		return nil
	})
	if err != nil {
		return false, err
	}
	if len(entries) == 0 {
		return false, nil
	}

	if err := p.enqueueBatchSideEffects(ctx, companyID, entries); err != nil {
		return false, err
	}

	hasMore := len(entries) >= p.batchSize
	if hasMore {
		if err := jobs.InsertBudgetProject(ctx, p.enqueuer, nil, companyID); err != nil {
			return false, err
		}
	}
	return hasMore, nil
}

func (p *Projector) enqueueBatchSideEffects(ctx context.Context, companyID int64, entries []types.UsageLedgerEntry) error {
	rebalanceAxes := make(map[string]struct{})
	touchedKeys := make(map[string]struct{})
	for _, entry := range entries {
		if entry.MemberID != nil {
			rebalanceAxes[store.RebalanceAxisMember+":"+*entry.MemberID] = struct{}{}
		}
		rebalanceAxes[store.RebalanceAxisDepartment+":"+entry.DepartmentID] = struct{}{}
		if entry.BudgetGroupID != nil {
			rebalanceAxes[store.RebalanceAxisBudgetGroup+":"+*entry.BudgetGroupID] = struct{}{}
		}
		if entry.PlatformKeyID != "" {
			touchedKeys[entry.PlatformKeyID] = struct{}{}
		}

		payload, err := json.Marshal(overrunPayload{
			DepartmentID:  entry.DepartmentID,
			MemberID:      entry.MemberID,
			BudgetGroupID: entry.BudgetGroupID,
			PlatformKeyID: entry.PlatformKeyID,
		})
		if err != nil {
			return err
		}
		if err := jobs.InsertOverrun(ctx, p.enqueuer, nil, companyID, payload); err != nil {
			return err
		}
	}
	for key := range rebalanceAxes {
		axisKind, axisID, ok := splitRebalanceKey(key)
		if !ok {
			continue
		}
		if err := jobs.InsertRebalance(ctx, p.enqueuer, nil, companyID, axisKind, axisID); err != nil {
			return err
		}
	}
	// Best-effort optional soft-block cache; skipped when disabled.
	budgetcheck.RefreshPlatformKeys(ctx, p.cfg, p.store, p.gatewayCache, p.logger, companyID, touchedKeys)
	return nil
}

func splitRebalanceKey(key string) (axisKind, axisID string, ok bool) {
	for i := 0; i < len(key); i++ {
		if key[i] == ':' {
			return key[:i], key[i+1:], true
		}
	}
	return "", "", false
}

func companyFromStore(co store.Company) company.Context {
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
