package newapisync

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/store"
)

func (l *NewAPISync) EnqueueRebalanceAxis(ctx context.Context, axisKind, axisID string) error {
	if !l.Enabled() {
		return nil
	}
	payload, _ := json.Marshal(RebalanceAxisOutboxPayload{
		CompanyID: company.CompanyID(ctx),
		AxisKind:  axisKind,
		AxisID:    axisID,
	})
	return l.outbox.EnqueueNewAPISyncOutbox(ctx, store.AsyncJob{
		ID:      fmt.Sprintf("outbox-%d", time.Now().UnixNano()),
		Kind:    store.OutboxKindRebalanceKey,
		Payload: payload,
		Status:  store.JobStatusPending,
	})
}
