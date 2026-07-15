package store

import (
	"context"
	"fmt"
	"unicode"
)

// IngestPendingChannel is the PostgreSQL NOTIFY channel used to wake the ingest worker.
const IngestPendingChannel = "ingest_pending"

// PGListener allows consumers to subscribe to PostgreSQL NOTIFY channels.
type PGListener interface {
	// WaitForNotification blocks until a notification arrives or ctx is cancelled.
	WaitForNotification(ctx context.Context) error
	// Listen subscribes to the given channel.
	Listen(ctx context.Context, channel string) error
	// Close releases the underlying connection.
	Close(ctx context.Context) error
}

// ValidPGNotifyChannel reports whether channel is a safe unquoted PostgreSQL identifier.
func ValidPGNotifyChannel(channel string) error {
	if channel == "" {
		return fmt.Errorf("empty notify channel")
	}
	for i, r := range channel {
		if i == 0 {
			if r != '_' && !unicode.IsLetter(r) {
				return fmt.Errorf("invalid notify channel %q", channel)
			}
			continue
		}
		if r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return fmt.Errorf("invalid notify channel %q", channel)
		}
	}
	return nil
}
