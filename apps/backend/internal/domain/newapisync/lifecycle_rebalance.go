package newapisync

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/company"
)

func (l *NewAPISync) EnqueueRebalanceAxis(ctx context.Context, axisKind, axisID string) error {
	if !l.Enabled() {
		return nil
	}
	return l.enqueuer.InsertRebalance(ctx, company.CompanyID(ctx), axisKind, axisID)
}
