package config

import (
	"fmt"
	"strings"
	"time"
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
	if anchor := strings.TrimSpace(c.ClockAnchor); anchor != "" {
		return anchor
	}
	return time.Now().UTC().Format("2006-01-02")
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
