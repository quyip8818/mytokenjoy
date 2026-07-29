//go:build testhook

package testutil

import "testing"

// SkipUnlessSaaS skips the test if not running in SaaS mode.
func SkipUnlessSaaS(t *testing.T) {
	t.Helper()
	if CurrentTestMode() != ModeSaaS {
		t.Skip("requires TEST_MODE=saas")
	}
}

// SkipUnlessLocal skips the test if not running in Local mode.
func SkipUnlessLocal(t *testing.T) {
	t.Helper()
	if CurrentTestMode() != ModeLocal {
		t.Skip("requires TEST_MODE=local")
	}
}
