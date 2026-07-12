package testutil

import (
	"context"
	"errors"

	"github.com/tokenjoy/backend/internal/domain/types"
)

// RecordingNotifier records all notifications sent for test assertions.
type RecordingNotifier struct {
	Notifications []types.Notification
}

func (n *RecordingNotifier) Send(_ context.Context, notification types.Notification) error {
	n.Notifications = append(n.Notifications, notification)
	return nil
}

// FailingNotifier always fails to send but does not return error (mimics fire-and-forget).
type FailingNotifier struct{}

func (n *FailingNotifier) Send(_ context.Context, _ types.Notification) error {
	return errors.New("notification delivery failed")
}
