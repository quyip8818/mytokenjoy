package config

import (
	"fmt"
	"strings"

	"github.com/tokenjoy/backend/internal/pkg/clock"
)

const (
	BootstrapNone    = "none"
	BootstrapMinimal = "minimal"
	BootstrapDemo    = "demo"
)

func (c Config) BootstrapIsNone() bool    { return c.BootstrapMode == BootstrapNone }
func (c Config) BootstrapIsMinimal() bool { return c.BootstrapMode == BootstrapMinimal }
func (c Config) BootstrapIsDemo() bool    { return c.BootstrapMode == BootstrapDemo }

func (c Config) SeedReferenceDate() string {
	return clock.NowUTC(c.Clock()).Format("2006-01-02")
}

func (c Config) DemoWithoutClockAnchor() bool {
	return c.BootstrapIsDemo() && strings.TrimSpace(c.ClockAnchor) == ""
}

func validateBootstrapMode(mode string) error {
	switch mode {
	case BootstrapNone, BootstrapMinimal, BootstrapDemo:
		return nil
	default:
		return fmt.Errorf("BOOTSTRAP_MODE must be none, minimal, or demo")
	}
}
