package worker

import (
	"context"
)

func (r *Runner) processRebalance(ctx context.Context) error {
	workerCtx := r.workerCtx(ctx, r.cfg.DefaultCompanyID)
	entries, err := r.rebalanceQueue.ClaimPendingRebalance(workerCtx, 20)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		entryCtx := r.workerCtx(ctx, entry.CompanyID)
		if err := r.rebalance.ProcessAxis(entryCtx, entry.AxisKind, entry.AxisID); err != nil {
			r.logger.Warn("rebalance failed", "axis", entry.AxisKind, "id", entry.AxisID, "company_id", entry.CompanyID, "error", err)
			if enqueueErr := r.lifecycle.EnqueueRebalanceAxis(entryCtx, entry.AxisKind, entry.AxisID); enqueueErr != nil {
				r.logger.Warn("re-enqueue rebalance failed", "axis", entry.AxisKind, "id", entry.AxisID, "error", enqueueErr)
			}
			continue
		}
		r.markRebalanceDone(workerCtx, entry.ID)
	}
	return nil
}
