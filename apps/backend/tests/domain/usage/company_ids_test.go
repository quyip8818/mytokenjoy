package usage_test

import (
	"testing"

	"github.com/tokenjoy/backend/seed/contract"
	"github.com/tokenjoy/backend/tests/testutil"
)

func TestCompanyIDsByLogID(t *testing.T) {
	t.Parallel()
	fix := newIngestFixture(t)

	testutil.SeedConsumeLog(t, fix.Store, testutil.DefaultConsumeLog(7801, 99))
	testutil.SeedConsumeLog(t, fix.Store, testutil.DefaultConsumeLog(7802, 99))
	testutil.SeedConsumeLog(t, fix.Store, testutil.DefaultConsumeLog(7803, 55))

	got, err := fix.Ingest.CompanyIDsByLogID(testutil.Ctx(), []int64{7801, 7802, 7803, 7804})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 resolved log IDs, got %d (%v)", len(got), got)
	}
	if got[7801] != contract.DefaultCompanyID || got[7802] != contract.DefaultCompanyID {
		t.Fatalf("unexpected mapping: %v", got)
	}
	if _, ok := got[7803]; ok {
		t.Fatal("unmapped token must be omitted")
	}
	if _, ok := got[7804]; ok {
		t.Fatal("missing log must be omitted")
	}
}

func TestCompanyIDsByLogIDEmpty(t *testing.T) {
	t.Parallel()
	fix := newIngestFixture(t, withoutMapping())
	got, err := fix.Ingest.CompanyIDsByLogID(testutil.Ctx(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}
