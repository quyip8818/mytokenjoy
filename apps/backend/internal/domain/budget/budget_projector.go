package budget

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/budget/schedule"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/store"
)

const defaultProjectorBatchSize = 500

type Projector struct {
	cfg              config.Config
	store            store.Store
	enqueuer         JobEnqueuer
	batchSize        int
	logger           *slog.Logger
	combinedKeyCache CombinedKeyCache
	notifier         types.Notifier
	// alertsSent tracks (ruleID:threshold:periodKey) to avoid repeat notifications.
	alertsSent map[string]struct{}
}

type batchEffects struct {
	touchedKeys     map[string]struct{}
	touchedDepts    map[string]struct{}
	keyIncrements   map[string]float64
	overrunByKey    map[string]overrunPayload
	deptOnlyOverrun *overrunPayload
	rebalanceAxes   map[string]struct{}
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

	if err := schedule.EnsureMonthRebalance(ctx, p.cfg, p.store, p.enqueuer, companyID); err != nil {
		return false, err
	}

	var entries []types.UsageLedgerEntry
	var summaries []store.CombinedKeySummary
	var effects batchEffects
	err = p.store.WithTx(ctx, func(tx store.Store) error {
		if err := tx.Budget().AcquireBudgetLock(ctx); err != nil {
			return err
		}
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
		effects = collectBatchEffects(batch)

		// Incremental path: decrement combined_key_remain by batch increments.
		// Absolute recompute only for keys that were not updated (e.g. combined_key_remain still NULL).
		if len(effects.keyIncrements) > 0 {
			summaries, err = tx.CombinedKeySummaries().DecrementBatch(ctx, effects.keyIncrements)
			if err != nil {
				return err
			}
			updated := make(map[string]struct{}, len(summaries))
			for _, item := range summaries {
				updated[item.PlatformKeyID] = struct{}{}
			}
			missing := make(map[string]struct{})
			for id := range effects.keyIncrements {
				if _, ok := updated[id]; !ok {
					missing[id] = struct{}{}
				}
			}
			if len(missing) > 0 {
				updates, err := ComputeGatewaySummaryUpdates(ctx, tx, missing, p.cfg.Clock())
				if err != nil {
					return err
				}
				if len(updates) > 0 {
					fallback, err := tx.CombinedKeySummaries().UpdateBatch(ctx, updates)
					if err != nil {
						return err
					}
					summaries = append(summaries, fallback...)
				}
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

	RefreshCombinedKeySummaries(ctx, p.combinedKeyCache, p.logger, companyID, summaries)
	if err := p.enqueueBatchEffects(ctx, companyID, effects); err != nil {
		return false, err
	}

	// Check percentage alert thresholds for touched departments.
	p.checkAlertThresholds(ctx, effects)

	hasMore := len(entries) >= p.batchSize
	if hasMore {
		if err := p.enqueuer.InsertBudgetProjection(ctx, companyID); err != nil {
			return false, err
		}
	}
	return hasMore, nil
}

func collectBatchEffects(entries []types.UsageLedgerEntry) batchEffects {
	effects := batchEffects{
		rebalanceAxes: make(map[string]struct{}),
		touchedKeys:   make(map[string]struct{}),
		touchedDepts:  make(map[string]struct{}),
		keyIncrements: make(map[string]float64),
		overrunByKey:  make(map[string]overrunPayload),
	}
	for _, entry := range entries {
		if entry.DepartmentID != "" {
			effects.touchedDepts[entry.DepartmentID] = struct{}{}
		}
		if entry.MemberID != nil && entry.PlatformKeyScope == types.PlatformKeyScopeMember {
			effects.rebalanceAxes[store.RebalanceAxisMember+":"+*entry.MemberID] = struct{}{}
		}
		if entry.ProjectID != nil {
			effects.rebalanceAxes[store.RebalanceAxisProject+":"+*entry.ProjectID] = struct{}{}
		}
		payload := overrunPayload{
			DepartmentID:  entry.DepartmentID,
			MemberID:      entry.MemberID,
			ProjectID:     entry.ProjectID,
			PlatformKeyID: entry.PlatformKeyID,
		}
		if entry.PlatformKeyID != "" {
			effects.touchedKeys[entry.PlatformKeyID] = struct{}{}
			effects.keyIncrements[entry.PlatformKeyID] += entry.Amount
			effects.overrunByKey[entry.PlatformKeyID] = payload
			continue
		}
		effects.deptOnlyOverrun = &payload
	}
	return effects
}

func (p *Projector) enqueueBatchEffects(ctx context.Context, companyID int64, effects batchEffects) error {
	for _, payload := range effects.overrunByKey {
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		if err := p.enqueuer.InsertOverrun(ctx, companyID, raw); err != nil {
			return err
		}
	}
	if effects.deptOnlyOverrun != nil {
		raw, err := json.Marshal(*effects.deptOnlyOverrun)
		if err != nil {
			return err
		}
		if err := p.enqueuer.InsertOverrun(ctx, companyID, raw); err != nil {
			return err
		}
	}
	for key := range effects.rebalanceAxes {
		axisKind, axisID, ok := splitRebalanceKey(key)
		if !ok {
			continue
		}
		if err := p.enqueuer.InsertRebalance(ctx, companyID, axisKind, axisID); err != nil {
			return err
		}
	}
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
