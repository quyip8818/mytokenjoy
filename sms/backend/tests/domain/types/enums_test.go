package types_test

import (
	"testing"

	"sms/backend/internal/domain/types"
)

func TestIsValidStatus(t *testing.T) {
	t.Parallel()
	cases := []struct {
		val     string
		allowed []string
		want    bool
	}{
		{"active", types.SupplierStatuses, true},
		{"frozen", types.SupplierStatuses, true},
		{"invalid", types.SupplierStatuses, false},
		{"pending", types.OrderStatuses, true},
		{"approved", types.OrderStatuses, true},
		{"unknown", types.OrderStatuses, false},
		{"A", types.EvalGrades, true},
		{"E", types.EvalGrades, false},
	}
	for _, tc := range cases {
		if got := types.IsValidStatus(tc.val, tc.allowed); got != tc.want {
			t.Errorf("IsValidStatus(%q, %v) = %v, want %v", tc.val, tc.allowed, got, tc.want)
		}
	}
}

func TestIsValidTransition(t *testing.T) {
	t.Parallel()
	cases := []struct {
		from, to string
		want     bool
	}{
		{"pending", "approved", true},
		{"pending", "cancelled", true},
		{"pending", "completed", false},
		{"approved", "delivered", true},
		{"approved", "cancelled", true},
		{"approved", "pending", false},
		{"delivered", "completed", true},
		{"delivered", "cancelled", true},
		{"delivered", "pending", false},
		{"completed", "pending", false}, // completed 无出口
		{"unknown", "pending", false},
	}
	for _, tc := range cases {
		if got := types.IsValidTransition(tc.from, tc.to); got != tc.want {
			t.Errorf("IsValidTransition(%q, %q) = %v, want %v", tc.from, tc.to, got, tc.want)
		}
	}
}
