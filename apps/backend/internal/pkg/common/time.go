package common

import (
	"fmt"
	"time"
)

func LoadLocation(tz string) (*time.Location, error) {
	if tz == "" {
		tz = "Asia/Shanghai"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone %q: %w", tz, err)
	}
	return loc, nil
}

func TruncateInTZ(t time.Time, unit time.Duration, loc *time.Location) time.Time {
	local := t.In(loc)
	switch unit {
	case 24 * time.Hour:
		y, m, d := local.Date()
		return time.Date(y, m, d, 0, 0, 0, 0, loc)
	case time.Hour:
		y, m, d := local.Date()
		return time.Date(y, m, d, local.Hour(), 0, 0, 0, loc)
	case time.Minute:
		y, m, d := local.Date()
		return time.Date(y, m, d, local.Hour(), local.Minute(), 0, 0, loc)
	default:
		return local
	}
}

func TruncateWeekInTZ(t time.Time, loc *time.Location) time.Time {
	local := t.In(loc)
	weekday := int(local.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	start := local.AddDate(0, 0, -(weekday - 1))
	y, m, d := start.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, loc)
}

func TruncateMonthInTZ(t time.Time, loc *time.Location) time.Time {
	local := t.In(loc)
	y, m, _ := local.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, loc)
}

func FormatBucketISO(t time.Time) string {
	return t.Format(time.RFC3339)
}
