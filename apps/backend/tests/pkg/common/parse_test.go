package common_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/pkg/common"
)

func TestParseIntParam(t *testing.T) {
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
