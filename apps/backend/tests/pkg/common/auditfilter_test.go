package common_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/pkg/common"
)

type sampleLog struct {
	CreatedAt string
	Detail    string
}

func TestFilterByDateRange(t *testing.T) {
	items := []sampleLog{
		{CreatedAt: "2026-06-01 10:00", Detail: "a"},
		{CreatedAt: "2026-06-15 10:00", Detail: "b"},
		{CreatedAt: "2026-06-30 10:00", Detail: "c"},
	}
	filtered := common.FilterByDateRangeCreatedAt(items, "2026-06-10", "2026-06-20", func(item sampleLog) string {
		return item.CreatedAt
	})
	if len(filtered) != 1 || filtered[0].Detail != "b" {
		t.Fatalf("expected one mid-month log, got %+v", filtered)
	}
}

func TestFilterByKeyword(t *testing.T) {
	items := []sampleLog{
		{CreatedAt: "2026-06-01", Detail: "create key"},
		{CreatedAt: "2026-06-02", Detail: "delete member"},
	}
	filtered := common.FilterByKeyword(items, "key", []func(sampleLog) string{
		func(item sampleLog) string { return item.Detail },
	})
	if len(filtered) != 1 || filtered[0].Detail != "create key" {
		t.Fatalf("expected keyword match on detail, got %+v", filtered)
	}
}
