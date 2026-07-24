package common_test

import (
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/pkg/common"
)

func TestParseIntParam(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		value    string
		fallback int
		want     int
	}{
		{"valid number", "5", 10, 5},
		{"empty string uses fallback", "", 10, 10},
		{"non-numeric uses fallback", "abc", 10, 10},
		{"zero uses fallback", "0", 10, 10},
		{"negative uses fallback", "-3", 10, 10},
		{"large number", "1000", 1, 1000},
		{"one", "1", 99, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := common.ParseIntParam(tt.value, tt.fallback)
			if got != tt.want {
				t.Errorf("ParseIntParam(%q, %d) = %d, want %d", tt.value, tt.fallback, got, tt.want)
			}
		})
	}
}

func TestTruncateInTZDayBoundary(t *testing.T) {
	t.Parallel()
	loc, err := common.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	ts := time.Date(2026, 6, 10, 15, 30, 0, 0, time.UTC)
	truncated := common.TruncateInTZ(ts, 24*time.Hour, loc)
	if truncated.Hour() != 0 || truncated.Location().String() != loc.String() {
		t.Fatalf("unexpected truncated time: %v", truncated)
	}
}
