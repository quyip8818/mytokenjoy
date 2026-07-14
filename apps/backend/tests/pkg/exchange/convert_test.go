package exchange_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/pkg/exchange"
)

func TestDefaultPPU(t *testing.T) {
	t.Parallel()
	if got := exchange.ToPoints(6.6); got != 6600 {
		t.Fatalf("ToPoints: got %v want 6600", got)
	}
	if got := exchange.ToDisplay(4950); got != 4.95 {
		t.Fatalf("ToDisplay: got %v want 4.95", got)
	}
}

func TestCustomPPU(t *testing.T) {
	t.Parallel()
	if got := exchange.ToPointsAt(1.5, 100); got != 150 {
		t.Fatalf("ToPointsAt: got %v want 150", got)
	}
	if got := exchange.ToDisplayAt(250, 100); got != 2.5 {
		t.Fatalf("ToDisplayAt: got %v want 2.5", got)
	}
}

func TestInvalidPPU(t *testing.T) {
	t.Parallel()
	if got := exchange.ToPointsAt(10, 0); got != 0 {
		t.Fatalf("ToPointsAt invalid ppu: got %v want 0", got)
	}
	if got := exchange.ToDisplayAt(10, 0); got != 0 {
		t.Fatalf("ToDisplayAt invalid ppu: got %v want 0", got)
	}
}
