//go:build testhook

package newapisync_test

import (
	"fmt"
	"testing"

	"github.com/tokenjoy/backend/internal/domain"
	domainnewapisync "github.com/tokenjoy/backend/internal/domain/newapisync"
)

func TestIsPermanentOutboxError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "service unavailable", err: domain.ServiceUnavailable("newapi not enabled"), want: true},
		{name: "platform key not found", err: fmt.Errorf("platform key not found"), want: false},
		{name: "transient", err: fmt.Errorf("upstream timeout"), want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := domainnewapisync.IsPermanentOutboxError(tc.err); got != tc.want {
				t.Fatalf("IsPermanentOutboxError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
