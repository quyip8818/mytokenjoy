package config

import (
	"fmt"
	"strings"

	"github.com/tokenjoy/backend/internal/pkg/clock"
)

const (
	BootstrapNone    = "none"
	BootstrapProd    = "prod"
	BootstrapMinimal = "minimal"
	BootstrapDemo    = "demo"
)

func (c Config) BootstrapIsNone() bool    { return c.BootstrapMode == BootstrapNone }
func (c Config) BootstrapIsProd() bool    { return c.BootstrapMode == BootstrapProd }
func (c Config) BootstrapIsMinimal() bool { return c.BootstrapMode == BootstrapMinimal }
func (c Config) BootstrapIsDemo() bool    { return c.BootstrapMode == BootstrapDemo }

// BootstrapNeedsSeed returns true when the mode requires data initialization.
func (c Config) BootstrapNeedsSeed() bool {
	return c.BootstrapIsProd() || c.BootstrapIsMinimal() || c.BootstrapIsDemo()
}

func (c Config) SeedReferenceDate() string {
	return clock.NowUTC(c.Clock()).Format("2006-01-02")
}

func (c Config) DemoWithoutClockAnchor() bool {
	return c.BootstrapIsDemo() && strings.TrimSpace(c.ClockAnchor) == ""
}

func validateBootstrapMode(mode string) error {
	switch mode {
	case BootstrapNone, BootstrapProd, BootstrapMinimal, BootstrapDemo:
		return nil
	default:
		return fmt.Errorf("BOOTSTRAP_MODE must be none, prod, minimal, or demo")
	}
}
