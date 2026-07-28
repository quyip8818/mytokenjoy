package store

import "context"

// SystemSettingsRepository is a global key-value store for system-level configuration
// (e.g. SMS sync partition versions). Not tenant-scoped.
type SystemSettingsRepository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
}
