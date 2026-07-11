package newapisync

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/infra/jobs"
)

func (l *NewAPISync) EnqueueRebalanceAxis(ctx context.Context, axisKind, axisID string) error {
	if !l.Enabled() {
		return nil
	}
	return jobs.InsertRebalance(ctx, l.enqueuer, nil, company.CompanyID(ctx), axisKind, axisID)
}
