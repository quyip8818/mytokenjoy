package worker

import (
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	tests := []struct {
		attempts int
		want     time.Duration
	}{
		{0, 1 * time.Second},    // 2^0 = 1
		{1, 2 * time.Second},    // 2^1 = 2
		{2, 4 * time.Second},    // 2^2 = 4
		{3, 8 * time.Second},    // 2^3 = 8
		{8, 256 * time.Second},  // 2^8 = 256
		{9, 300 * time.Second},  // 2^9 = 512, capped at 300
		{20, 300 * time.Second}, // far beyond cap
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := backoff(tt.attempts)
			if got != tt.want {
				t.Errorf("backoff(%d) = %v, want %v", tt.attempts, got, tt.want)
			}
		})
	}
}
