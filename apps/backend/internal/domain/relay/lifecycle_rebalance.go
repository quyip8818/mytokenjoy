package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/store"
)

func (l *TokenLifecycle) EnqueueRebalanceAxis(ctx context.Context, axisKind, axisID string) error {
	if !l.Enabled() {
		return nil
	}
	payload, _ := json.Marshal(RebalanceAxisOutboxPayload{
		CompanyID: company.CompanyID(ctx),
		AxisKind:  axisKind,
		AxisID:    axisID,
	})
	return l.relayOutbox.EnqueueRelayOutbox(ctx, store.RelayOutboxEntry{
		ID:        fmt.Sprintf("outbox-%d", time.Now().UnixNano()),
		Kind:      store.OutboxKindRebalanceToken,
		Payload:   payload,
		Status:    store.OutboxStatusPending,
		NextRetry: time.Now(),
		CreatedAt: time.Now(),
	})
}
