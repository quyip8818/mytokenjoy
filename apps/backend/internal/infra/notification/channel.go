package notification

import (
	"context"

	"github.com/google/uuid"
	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
)

// Channel is the interface each delivery channel must implement.
type Channel interface {
	// Name returns the channel identifier (e.g. "email", "sms", "in_app").
	Name() string
	// IsConfigured returns true if this channel has all required configuration to operate.
	IsConfigured() bool
	// Send delivers a rendered message to the given recipient.
	Send(ctx context.Context, recipientID uuid.UUID, msg domainnotification.RenderedMessage) error
}

// DirectSender is an optional interface a Channel can implement to support
// sending directly to an address (phone/email) without recipient ID resolution.
// Used by verification code flows where the recipient may not exist yet.
type DirectSender interface {
	// SendDirect delivers a rendered message directly to the given address.
	SendDirect(ctx context.Context, address string, msg domainnotification.RenderedMessage) error
}
