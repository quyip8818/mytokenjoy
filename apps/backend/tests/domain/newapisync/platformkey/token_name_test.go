package platformkey_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/domain/newapisync/platformkey"
)

func TestTokenNameUniquePerPlatformKey(t *testing.T) {
	t.Parallel()
	a := platformkey.TokenName("plk-1")
	b := platformkey.TokenName("plk-2")
	if a == b {
		t.Fatalf("expected distinct names, both %q", a)
	}
	if a != "tokenjoy:plk-1" {
		t.Fatalf("unexpected name %q", a)
	}
}
