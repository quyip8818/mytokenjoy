package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

// InAppChannel writes notifications to the notification_log table for user inbox display.
// It is always configured since it only requires the database (no external provider).
type InAppChannel struct {
	store  store.Store
	logger *slog.Logger
	hub    *SSEHub // optional: for real-time push
}

func NewInAppChannel(st store.Store, logger *slog.Logger, hub *SSEHub) *InAppChannel {
	return &InAppChannel{
		store:  st,
		logger: logger,
		hub:    hub,
	}
}

func (c *InAppChannel) Name() string { return domainnotification.ChannelInApp }

func (c *InAppChannel) IsConfigured() bool { return true }

func (c *InAppChannel) Send(ctx context.Context, recipientID uuid.UUID, msg domainnotification.RenderedMessage) error {
	groupKey := extractString(msg.Payload, "_groupKey")
	category := extractString(msg.Payload, "_category")

	// Strip internal metadata fields before persisting payload
	cleanPayload := make(map[string]any, len(msg.Payload))
	for k, v := range msg.Payload {
		if k == "_groupKey" || k == "_category" {
			continue
		}
		cleanPayload[k] = v
	}

	payload, err := json.Marshal(cleanPayload)
	if err != nil {
		payload = []byte("{}")
	}

	entry := types.NotificationLogEntry{
		ID:        uuid.Must(uuid.NewV7()),
		Channel:   domainnotification.ChannelInApp,
		EventType: extractEventType(msg.Payload),
		UserID:    recipientID,
		Title:     msg.Title,
		Body:      msg.Body,
		Payload:   payload,
		SendOK:    true,
		Category:  category,
		GroupKey:  groupKey,
	}

	if err := c.store.Notification().Append(ctx, entry); err != nil {
		return fmt.Errorf("in_app channel write: %w", err)
	}

	// Push to SSE hub if available
	if c.hub != nil && recipientID != uuid.Nil {
		c.hub.Publish(recipientID, SSEEvent{
			ID:        entry.ID.String(),
			EventType: entry.EventType,
			Title:     msg.Title,
			Body:      msg.Body,
			GroupKey:  groupKey,
			Category:  category,
			Payload:   payload,
		})
	}

	return nil
}

func extractEventType(payload map[string]any) string {
	if payload == nil {
		return "unknown"
	}
	if v, ok := payload["eventType"].(string); ok && v != "" {
		return v
	}
	return "unknown"
}

var _ Channel = (*InAppChannel)(nil)
