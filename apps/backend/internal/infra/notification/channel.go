package notification

import (
	"context"

	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
)

// Channel is the interface each delivery channel must implement.
type Channel interface {
	// Name returns the channel identifier (e.g. "email", "sms", "in_app").
	Name() string
	// IsConfigured returns true if this channel has all required configuration to operate.
	IsConfigured() bool
	// Send delivers a rendered message to the given recipient.
	Send(ctx context.Context, recipientID string, msg domainnotification.RenderedMessage) error
}
