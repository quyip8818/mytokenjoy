package bootstrap

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds bootstrap initialization parameters.
// Loaded from BOOTSTRAP_CONFIG_PATH or defaults.
type Config struct {
	Version int           `yaml:"version"`
	Company CompanyConfig `yaml:"company"`
	Admin   AdminConfig   `yaml:"admin"`
	Billing BillingConfig `yaml:"billing"`
	Models  []ModelConfig `yaml:"models"`
}

type CompanyConfig struct {
	Name string `yaml:"name"`
}

type AdminConfig struct {
	Name     string `yaml:"name"`
	Email    string `yaml:"email"`
	Password string `yaml:"password"`
}

type BillingConfig struct {
	Currency     string `yaml:"currency"`
	QuotaPerUnit int64  `yaml:"quota_per_unit"`
}

type ModelConfig struct {
	CallType    string  `yaml:"call_type"`
	Name        string  `yaml:"name"`
	Provider    string  `yaml:"provider"`
	InputRatio  float64 `yaml:"input_ratio"`
	OutputRatio float64 `yaml:"output_ratio"`
}

// DefaultConfig returns the built-in default bootstrap configuration.
func DefaultConfig() Config {
	return Config{
		Version: 1,
		Company: CompanyConfig{Name: "My Company"},
		Billing: BillingConfig{
			Currency:     "CNY",
			QuotaPerUnit: 500_000,
		},
	}
}

// LoadConfig reads a bootstrap config from path. If path is empty, returns DefaultConfig.
func LoadConfig(path string) (Config, error) {
	if path == "" {
		return DefaultConfig(), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read bootstrap config %s: %w", path, err)
	}
	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse bootstrap config %s: %w", path, err)
	}
	if cfg.Company.Name == "" {
		return Config{}, fmt.Errorf("bootstrap config: company.name is required")
	}
	if cfg.Billing.QuotaPerUnit <= 0 {
		return Config{}, fmt.Errorf("bootstrap config: billing.quota_per_unit must be positive")
	}
	return cfg, nil
}
