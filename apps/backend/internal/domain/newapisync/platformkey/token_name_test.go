package platformkey

import "testing"

func TestTokenNameUniquePerPlatformKey(t *testing.T) {
	t.Parallel()
	a := TokenName("plk-1")
	b := TokenName("plk-2")
	if a == b {
		t.Fatalf("expected distinct names, both %q", a)
	}
	if a != "tokenjoy:plk-1" {
		t.Fatalf("unexpected name %q", a)
	}
}
