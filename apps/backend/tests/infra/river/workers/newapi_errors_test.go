package workers_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/infra/river/workers"
)

func TestIsNonRetryableNewAPIError(t *testing.T) {
	t.Parallel()
	cases := []struct {
		err  error
		want bool
	}{
		{nil, false},
		{errors.New("temporary network blip"), false},
		{fmt.Errorf("newapi: ERROR: bigint out of range (SQLSTATE 22003)"), true},
		{fmt.Errorf("wrap: %w", fmt.Errorf("sqlstate 22003")), true},
		{fmt.Errorf("topup quota delta out of range"), true},
		{fmt.Errorf("newapi wallet user id required for platform key x"), true},
		{domain.ServiceUnavailable("newapi disabled"), true},
	}
	for _, tc := range cases {
		if got := workers.IsNonRetryableNewAPIError(tc.err); got != tc.want {
			t.Fatalf("IsNonRetryableNewAPIError(%v)=%v want %v", tc.err, got, tc.want)
		}
	}
}
