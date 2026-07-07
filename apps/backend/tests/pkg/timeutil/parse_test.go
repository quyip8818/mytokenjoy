package timeutil_test

import (
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/pkg/timeutil"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{"RFC3339", "2024-03-15T10:30:00Z", time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC), false},
		{"RFC3339 with offset", "2024-03-15T10:30:00+08:00", time.Date(2024, 3, 15, 10, 30, 0, 0, time.FixedZone("", 8*3600)), false},
		{"date and time with minutes", "2024-03-15 10:30", time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC), false},
		{"date and time with seconds", "2024-03-15 10:30:45", time.Date(2024, 3, 15, 10, 30, 45, 0, time.UTC), false},
		{"date only", "2024-03-15", time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC), false},
		{"empty string", "", time.Time{}, true},
		{"invalid format", "not-a-date", time.Time{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := timeutil.Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && !got.Equal(tt.want) {
				t.Errorf("Parse(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseOrNow(t *testing.T) {
	t.Run("empty returns now", func(t *testing.T) {
		before := time.Now().UTC()
		got, err := timeutil.ParseOrNow("")
		after := time.Now().UTC()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Before(before.Add(-time.Second)) || got.After(after.Add(time.Second)) {
			t.Errorf("expected time near now, got %v", got)
		}
	})

	t.Run("valid value parses", func(t *testing.T) {
		got, err := timeutil.ParseOrNow("2024-06-01")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
		if !got.Equal(want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("invalid value errors", func(t *testing.T) {
		_, err := timeutil.ParseOrNow("bad")
		if err == nil {
			t.Fatal("expected error for invalid input")
		}
	})
}

func TestFormatSyncLog(t *testing.T) {
	input := time.Date(2024, 3, 15, 14, 30, 45, 0, time.UTC)
	got := timeutil.FormatSyncLog(input)
	want := "2024-03-15 14:30"
	if got != want {
		t.Errorf("FormatSyncLog = %q, want %q", got, want)
	}
}

func TestFormatDateOnly(t *testing.T) {
	input := time.Date(2024, 12, 25, 23, 59, 59, 0, time.UTC)
	got := timeutil.FormatDateOnly(input)
	want := "2024-12-25"
	if got != want {
		t.Errorf("FormatDateOnly = %q, want %q", got, want)
	}
}
