package worker

import (
	"context"
)

func (r *Runner) processOverrun(ctx context.Context) error {
	workerCtx := r.workerCtx(ctx, r.cfg.DefaultCompanyID)
	entries, err := r.asyncJobs.ClaimPendingOverrun(workerCtx, 20)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		entryCtx := r.workerCtx(ctx, entry.CompanyID)
		if err := r.overrun.ProcessOverrunPayload(entryCtx, entry.Payload); err != nil {
			r.logger.Warn("overrun failed", "id", entry.ID, "company_id", entry.CompanyID, "error", err)
			continue
		}
		r.markOverrunDone(workerCtx, entry.ID)
	}
	return nil
}

func (r *Runner) markOverrunDone(ctx context.Context, id string) {
	if err := r.asyncJobs.MarkOverrunDone(ctx, id); err != nil {
		r.logger.Warn("mark overrun done failed", "id", id, "error", err)
	}
}
