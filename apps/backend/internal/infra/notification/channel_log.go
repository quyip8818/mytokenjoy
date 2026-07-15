package notification

import (
	"context"
	"log/slog"

	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
)

// LogChannel logs notifications to the application logger. Always configured.
type LogChannel struct {
	logger *slog.Logger
}

func NewLogChannel(logger *slog.Logger) *LogChannel {
	return &LogChannel{logger: logger}
}

func (c *LogChannel) Name() string { return domainnotification.ChannelLog }

func (c *LogChannel) IsConfigured() bool { return true }

func (c *LogChannel) Send(ctx context.Context, recipientID string, msg domainnotification.RenderedMessage) error {
	c.logger.Info("notification",
		"channel", "log",
		"recipient", recipientID,
		"title", msg.Title,
		"body", msg.Body,
	)
	return nil
}

var _ Channel = (*LogChannel)(nil)
