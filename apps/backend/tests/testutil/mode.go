//go:build testhook

package testutil

import "os"

// TestMode represents the deployment mode used for testing.
type TestMode string

const (
	ModeSaaS  TestMode = "saas"
	ModeLocal TestMode = "local"
)

// CurrentTestMode returns the test mode from TEST_MODE env. Defaults to ModeSaaS.
func CurrentTestMode() TestMode {
	if os.Getenv("TEST_MODE") == "local" {
		return ModeLocal
	}
	return ModeSaaS
}
