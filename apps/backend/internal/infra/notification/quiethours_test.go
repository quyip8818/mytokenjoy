package notification

import (
	"testing"

	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
)

func TestIsInQuietHoursNil(t *testing.T) {
	t.Parallel()
	if IsInQuietHours(nil) {
		t.Fatal("nil quiet hours should return false")
	}
}

func TestIsInQuietHoursEmpty(t *testing.T) {
	t.Parallel()
	qh := &domainnotification.QuietHours{Start: "", End: "", Timezone: "UTC"}
	if IsInQuietHours(qh) {
		t.Fatal("empty start/end should return false")
	}
}

func TestIsInQuietHoursInvalidTimezone(t *testing.T) {
	t.Parallel()
	qh := &domainnotification.QuietHours{Start: "22:00", End: "08:00", Timezone: "Invalid/TZ"}
	if IsInQuietHours(qh) {
		t.Fatal("invalid timezone should return false")
	}
}

func TestParseTimeMinutesValid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input    string
		expected int
	}{
		{"00:00", 0},
		{"08:30", 510},
		{"23:59", 1439},
		{"12:00", 720},
	}
	for _, tc := range cases {
		got := parseTimeMinutes(tc.input)
		if got != tc.expected {
			t.Errorf("parseTimeMinutes(%q) = %d, want %d", tc.input, got, tc.expected)
		}
	}
}

func TestParseTimeMinutesInvalid(t *testing.T) {
	t.Parallel()
	invalids := []string{"", "8:00", "abc", "25:00", "12:60", "12-00"}
	for _, s := range invalids {
		if parseTimeMinutes(s) != -1 {
			t.Errorf("parseTimeMinutes(%q) should return -1", s)
		}
	}
}
