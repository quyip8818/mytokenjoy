package config

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port          string `env:"PORT" envDefault:"8080"`
	CORSOrigins   string `env:"CORS_ORIGINS" envDefault:"http://localhost:5173"`
	SimulateDelay bool   `env:"SIMULATE_DELAY" envDefault:"true"`
	DemoToday     string `env:"DEMO_TODAY" envDefault:"2026-06-19"`
}

func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

func (c Config) CORSOriginList() []string {
	parts := strings.Split(c.CORSOrigins, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}
