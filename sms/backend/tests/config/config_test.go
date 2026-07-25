package config_test

import (
	"testing"

	"sms/backend/internal/config"
)

func TestNewAPIEnabled(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		baseURL string
		want    bool
	}{
		{"empty disables", "", false},
		{"set enables", "http://localhost:3000", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.Config{NewAPIBaseURL: tc.baseURL}
			if got := cfg.NewAPIEnabled(); got != tc.want {
				t.Fatalf("NewAPIEnabled() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNewAPIDBURL_Derivation(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		dbURL       string
		explicitURL string
		want        string
	}{
		{
			"derives from DATABASE_URL with params",
			"postgres://sms:sms@localhost:5432/sms?sslmode=disable",
			"",
			"postgres://sms:sms@localhost:5432/newapi?sslmode=disable",
		},
		{
			"derives from DATABASE_URL without params",
			"postgres://sms:sms@localhost:5432/sms",
			"",
			"postgres://sms:sms@localhost:5432/newapi",
		},
		{
			"explicit overrides derivation",
			"postgres://sms:sms@localhost:5432/sms?sslmode=disable",
			"postgres://other:pass@host:5433/newapi",
			"postgres://other:pass@host:5433/newapi",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.Config{
				DatabaseURL:       tc.dbURL,
				NewAPIDatabaseURL: tc.explicitURL,
			}
			if got := cfg.NewAPIDBURL(); got != tc.want {
				t.Fatalf("NewAPIDBURL() = %q, want %q", got, tc.want)
			}
		})
	}
}
