package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port              string `env:"PORT" envDefault:"8080"`
	DatabaseURL       string `env:"DATABASE_URL,required"`
	CORSOrigins       string `env:"CORS_ORIGINS" envDefault:"http://localhost:5173"`
	JWTSecret         string `env:"JWT_SECRET" envDefault:"sms-dev-secret"`
	AccessTokenTTLMin int    `env:"ACCESS_TOKEN_TTL_MIN" envDefault:"15"`
	RefreshTokenTTLH  int    `env:"REFRESH_TOKEN_TTL_H" envDefault:"168"`
	UploadDir         string `env:"UPLOAD_DIR" envDefault:"./uploads"`
	SecureCookie      bool   `env:"SECURE_COOKIE" envDefault:"false"`
}

func (c Config) CORSOriginList() []string {
	parts := strings.Split(c.CORSOrigins, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			origins = append(origins, t)
		}
	}
	return origins
}

func (c Config) AccessTokenTTL() time.Duration {
	return time.Duration(c.AccessTokenTTLMin) * time.Minute
}

func (c Config) RefreshTokenTTL() time.Duration {
	return time.Duration(c.RefreshTokenTTLH) * time.Hour
}

func Load() (Config, error) {
	loadDotEnv(".env.development")
	loadDotEnv(".env")

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

func loadDotEnv(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k, v = strings.TrimSpace(k), strings.TrimSpace(v)
		if _, exists := os.LookupEnv(k); !exists {
			os.Setenv(k, v)
		}
	}
}
