package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

type Notifier interface {
	Send(ctx context.Context, notification types.Notification) error
}

type Service struct {
	cfg        config.Config
	store      store.Store
	logger     *slog.Logger
	httpClient *http.Client
}

func NewService(cfg config.Config, st store.Store, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		cfg:        cfg,
		store:      st,
		logger:     logger,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

var _ Notifier = (*Service)(nil)

func (s *Service) Send(ctx context.Context, notification types.Notification) error {
	payload, err := json.Marshal(notification.Payload)
	if err != nil {
		return err
	}
	entry := types.NotificationLogEntry{
		ID:        fmt.Sprintf("ntf-%d", time.Now().UnixNano()),
		Channel:   types.NotificationChannelLog,
		EventType: notification.EventType,
		Recipient: notification.Recipient,
		Payload:   payload,
		Status:    types.NotificationStatusSent,
	}
	s.logger.Info(
		"notification",
		"event", notification.EventType,
		"recipient", notification.Recipient,
		"payload", string(payload),
	)
	if err := s.store.Notification().Append(ctx, entry); err != nil {
		return err
	}
	if strings.TrimSpace(s.cfg.NotifyWebhookURL) == "" {
		return nil
	}
	return s.sendWebhook(ctx, notification, payload)
}

func (s *Service) sendWebhook(ctx context.Context, notification types.Notification, payload []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.NotifyWebhookURL, strings.NewReader(string(payload)))
	if err != nil {
		s.recordWebhookFailure(ctx, notification, payload, err)
		return nil
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Event-Type", notification.EventType)
	res, err := s.httpClient.Do(req)
	if err != nil {
		s.recordWebhookFailure(ctx, notification, payload, err)
		return nil
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, res.Body)
	if res.StatusCode >= 400 {
		s.recordWebhookFailure(ctx, notification, payload, fmt.Errorf("webhook status %d", res.StatusCode))
		return nil
	}
	webhookEntry := types.NotificationLogEntry{
		ID:        fmt.Sprintf("ntf-%d", time.Now().UnixNano()),
		Channel:   types.NotificationChannelWebhook,
		EventType: notification.EventType,
		Recipient: s.cfg.NotifyWebhookURL,
		Payload:   payload,
		Status:    types.NotificationStatusSent,
	}
	_ = s.store.Notification().Append(ctx, webhookEntry)
	return nil
}

func (s *Service) recordWebhookFailure(
	ctx context.Context,
	notification types.Notification,
	payload []byte,
	sendErr error,
) {
	entry := types.NotificationLogEntry{
		ID:        fmt.Sprintf("ntf-%d", time.Now().UnixNano()),
		Channel:   types.NotificationChannelWebhook,
		EventType: notification.EventType,
		Recipient: s.cfg.NotifyWebhookURL,
		Payload:   payload,
		Status:    types.NotificationStatusFailed,
		Error:     sendErr.Error(),
	}
	_ = s.store.Notification().Append(ctx, entry)
	s.logger.Warn("notification webhook failed", "event", notification.EventType, "error", sendErr)
}

func LogSyncThresholdExceeded(logger *slog.Logger, cfg types.SyncConfig, detail string) {
	if logger == nil {
		logger = slog.Default()
	}
	logger.Warn(
		"sync threshold exceeded",
		"detail", detail,
		"notifyPhone", cfg.NotifyPhone,
		"notifyEmail", cfg.NotifyEmail,
		"notifyIm", cfg.NotifyIm,
	)
}

func NotifySyncThresholdExceeded(ctx context.Context, notifier Notifier, cfg types.SyncConfig, detail string) {
	if notifier == nil {
		LogSyncThresholdExceeded(slog.Default(), cfg, detail)
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
