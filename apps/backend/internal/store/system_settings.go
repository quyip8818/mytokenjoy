package store

import "context"

// SystemSettingsRepository is a global key-value store for system-level configuration
// (e.g. catalog sync versions). Not tenant-scoped.
type SystemSettingsRepository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
	// Increment atomically increments a numeric key by 1 and returns the new value.
	Increment(ctx context.Context, key string) (int, error)
}
