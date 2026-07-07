package common_test

import (
	"context"
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/pkg/common"
)

func TestDelayerDisabled(t *testing.T) {
	d := common.NewDelayer(false)
	start := time.Now()
	err := d.Wait(context.Background(), 5*time.Second)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elapsed > 100*time.Millisecond {
		t.Errorf("disabled delayer should not wait, took %v", elapsed)
	}
}

func TestDelayerEnabled(t *testing.T) {
	d := common.NewDelayer(true)
	start := time.Now()
	err := d.Wait(context.Background(), 50*time.Millisecond)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elapsed < 40*time.Millisecond {
		t.Errorf("enabled delayer should wait, only took %v", elapsed)
	}
}

func TestDelayerCancelledContext(t *testing.T) {
	d := common.NewDelayer(true)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := d.Wait(ctx, 5*time.Second)
	if err == nil {
		t.Fatal("expected context error")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}
