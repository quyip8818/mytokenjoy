package usage_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/types"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
)

func TestValidateCostGranularity(t *testing.T) {
	if err := domainusage.ValidateCostGranularity(types.UsageGranularityWeek); err != nil {
		t.Fatalf("expected week to be valid: %v", err)
	}
	if err := domainusage.ValidateCostGranularity(""); err != nil {
		t.Fatalf("expected empty granularity to be valid: %v", err)
	}
	if err := domainusage.ValidateCostGranularity(types.UsageGranularityMinute); err == nil {
		t.Fatal("expected minute to be invalid for cost endpoints")
	}
}

func TestNormalizeCostGranularity(t *testing.T) {
	if got := domainusage.NormalizeCostGranularity(""); got != types.UsageGranularityDay {
		t.Fatalf("expected day default, got %s", got)
	}
}
