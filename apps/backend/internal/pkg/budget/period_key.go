package budget

import "time"

const PeriodMonthly = "monthly"

// SnapshotKey resolves a period_key string from an org period spec and an instant.
// Fixed specs (e.g. "2026-06") are returned as-is; empty / PeriodMonthly use at's UTC month.
// Domain open-budget paths must use Open* factories; this is the shared string primitive for
// store/seed and internal resolution.
func SnapshotKey(orgPeriod string, at time.Time) string {
	if orgPeriod != "" && orgPeriod != PeriodMonthly {
		return orgPeriod
	}
	return at.UTC().Format("2006-01")
}
