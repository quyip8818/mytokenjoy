package remote

import (
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/clock"
)

func ComputeNextOrgSync(cfg types.SyncConfig, lastSyncAt *time.Time, clk clock.Clock) time.Time {
	now := clock.NowUTC(clk)
	if !cfg.Enabled || cfg.FrequencyHours < 1 {
		return now.UTC()
	}

	interval := time.Duration(cfg.FrequencyHours) * time.Hour
	anchor := now
	if lastSyncAt != nil {
		anchor = lastSyncAt.Add(interval)
	}
	if cfg.StartTime == "" {
		if anchor.Before(now) {
			return now.UTC()
		}
		return anchor.UTC()
	}

	parsed, err := time.Parse("15:04", cfg.StartTime)
	if err != nil {
		if anchor.Before(now) {
			return now.UTC()
		}
		return anchor.UTC()
	}

	loc := now.Location()
	candidate := time.Date(anchor.Year(), anchor.Month(), anchor.Day(), parsed.Hour(), parsed.Minute(), 0, 0, loc)
	for candidate.Before(anchor) {
		candidate = candidate.Add(24 * time.Hour)
	}
	for candidate.Before(now) {
		candidate = candidate.Add(interval)
	}
	return candidate.UTC()
}
