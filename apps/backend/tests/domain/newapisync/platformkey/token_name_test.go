package platformkey_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/newapisync/platformkey"
)

func TestTokenNameUniquePerPlatformKey(t *testing.T) {
	t.Parallel()
	idA := uuid.MustParse("00000000-0000-7000-0000-00000000f001")
	idB := uuid.MustParse("00000000-0000-7000-0000-00000000f002")
	a := platformkey.TokenName(idA)
	b := platformkey.TokenName(idB)
	if a == b {
		t.Fatalf("expected distinct names, both %q", a)
	}
	if a != "tokenjoy:00000000-0000-7000-0000-00000000f001" {
		t.Fatalf("unexpected name %q", a)
	}
}
