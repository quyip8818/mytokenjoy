package common_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/pkg/common"
)

func TestPaginate(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	tests := []struct {
		name         string
		page         int
		pageSize     int
		wantResult   []int
		wantTotal    int
		wantPage     int
		wantPageSize int
	}{
		{"first page", 1, 3, []int{1, 2, 3}, 10, 1, 3},
		{"second page", 2, 3, []int{4, 5, 6}, 10, 2, 3},
		{"last partial page", 4, 3, []int{10}, 10, 4, 3},
		{"page beyond range", 5, 3, []int{}, 10, 5, 3},
		{"zero page defaults to 1", 0, 3, []int{1, 2, 3}, 10, 1, 3},
		{"negative page defaults to 1", -1, 3, []int{1, 2, 3}, 10, 1, 3},
		{"zero pageSize defaults to 1", 1, 0, []int{1}, 10, 1, 1},
		{"negative pageSize defaults to 1", 1, -5, []int{1}, 10, 1, 1},
		{"full page", 1, 10, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 10, 1, 10},
		{"pageSize larger than total", 1, 20, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 10, 1, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, total, safePage, safeSize := common.Paginate(items, tt.page, tt.pageSize)
			if total != tt.wantTotal {
				t.Errorf("total = %d, want %d", total, tt.wantTotal)
			}
			if safePage != tt.wantPage {
				t.Errorf("safePage = %d, want %d", safePage, tt.wantPage)
			}
			if safeSize != tt.wantPageSize {
				t.Errorf("safeSize = %d, want %d", safeSize, tt.wantPageSize)
			}
			if len(result) != len(tt.wantResult) {
				t.Fatalf("result len = %d, want %d", len(result), len(tt.wantResult))
			}
			for i, v := range result {
				if v != tt.wantResult[i] {
					t.Errorf("result[%d] = %d, want %d", i, v, tt.wantResult[i])
				}
			}
		})
	}
}

func TestPaginateEmptySlice(t *testing.T) {
	result, total, _, _ := common.Paginate([]string{}, 1, 10)
	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
	if len(result) != 0 {
		t.Errorf("result len = %d, want 0", len(result))
	}
}
