package timeutil

import (
	"fmt"
	"time"
)

func Parse(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, fmt.Errorf("empty time")
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02T15:04:05Z07:00", value); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02 15:04", value); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02 15:04:05", value); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02", value); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("unsupported time format: %q", value)
}

func ParseOrNow(value string) (time.Time, error) {
	if value == "" {
		return time.Now().UTC(), nil
	}
	return Parse(value)
}

func FormatSyncLog(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}

func FormatDateOnly(t time.Time) string {
	return t.Format("2006-01-02")
}
