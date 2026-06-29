package common_test

import (
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/pkg/common"
)

func TestTruncateInTZDayBoundary(t *testing.T) {
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
