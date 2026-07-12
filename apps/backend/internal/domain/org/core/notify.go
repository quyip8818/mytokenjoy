package core

import (
	"context"

	"github.com/tokenjoy/backend/internal/domain/types"
)

func NotifySyncThresholdExceeded(ctx context.Context, notifier Notifier, cfg types.SyncConfig, detail string) {
	if notifier == nil {
		return
	}
	_ = notifier.Send(ctx, types.Notification{
		EventType: types.NotificationEventSyncThreshold,
		Payload: map[string]any{
			"detail":      detail,
			"notifyPhone": cfg.NotifyPhone,
			"notifyEmail": cfg.NotifyEmail,
			"notifyIm":    cfg.NotifyIm,
		},
	})
}
