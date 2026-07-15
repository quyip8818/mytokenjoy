package ingest

import (
	"reflect"
	"testing"

	"github.com/tokenjoy/backend/internal/store"
)

func TestGroupJobsByCompany(t *testing.T) {
	t.Parallel()
	jobs := []store.IngestJob{
		{ID: "a", LogID: 1},
		{ID: "b", LogID: 2},
		{ID: "c", LogID: 3},
		{ID: "d", LogID: 4},
	}
	companyByLogID := map[int64]int64{
		1: 10,
		2: 20,
		3: 10,
	}
	got := groupJobsByCompany(jobs, companyByLogID)
	want := [][]store.IngestJob{
		{{ID: "a", LogID: 1}, {ID: "c", LogID: 3}},
		{{ID: "b", LogID: 2}},
		{{ID: "d", LogID: 4}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestGroupJobsByCompanyEmpty(t *testing.T) {
	t.Parallel()
	if got := groupJobsByCompany(nil, nil); got != nil {
		t.Fatalf("expected nil, got %#v", got)
	}
}
