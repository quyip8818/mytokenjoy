package periodutil_test

import (
	"testing"
	"time"

	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/pkg/periodutil"
)

func TestResolveCurrentMonth(t *testing.T) {
	now := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)
	rng, err := periodutil.Resolve(types.CostQueryParams{Period: string(types.CostPeriodCurrentMonth)}, now, domainusage.DefaultTimezone)
	if err != nil {
		t.Fatal(err)
	}
	if rng.Timezone != domainusage.DefaultTimezone {
		t.Fatalf("expected timezone %s, got %s", domainusage.DefaultTimezone, rng.Timezone)
	}
	if rng.End.Sub(rng.Start) < 28*24*time.Hour {
		t.Fatalf("unexpected current month range: %+v", rng)
	}
}

func TestPreviousRange(t *testing.T) {
	current := periodutil.ResolvedRange{
		Start: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	prev := periodutil.PreviousRange(current)
	if !prev.End.Equal(current.Start) {
		t.Fatalf("expected previous end at current start, got %+v", prev)
	}
}
