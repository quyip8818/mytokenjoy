package types

import "context"

// Notifier sends domain notifications without coupling to infra/notification.
type Notifier interface {
	Send(ctx context.Context, notification Notification) error
}
