package newapisync

import "testing"

func TestNewAPITokenNameUniquePerPlatformKey(t *testing.T) {
	t.Parallel()
	a := newAPITokenName("plk-1")
	b := newAPITokenName("plk-2")
	if a == b {
		t.Fatalf("expected distinct names, both %q", a)
	}
	if a != "tokenjoy:plk-1" {
		t.Fatalf("unexpected name %q", a)
	}
}
