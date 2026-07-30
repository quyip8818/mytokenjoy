package config

import "github.com/tokenjoy/backend/internal/pkg/clock"

// SeedReferenceDate returns the current date string used by seed/snapshot as the reference date.
func (c Config) SeedReferenceDate() string {
	return clock.NowUTC(c.Clock()).Format("2006-01-02")
}
