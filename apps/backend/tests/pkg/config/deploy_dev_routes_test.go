package config_test

import (
	"testing"

	"github.com/tokenjoy/backend/internal/config"
)

func TestAllowsDevHTTPRoutesOnlyLocal(t *testing.T) {
	t.Parallel()
	cases := []struct {
		env  string
		want bool
	}{
		{config.DeployEnvLocal, true},
		{config.DeployEnvStaging, false},
		{config.DeployEnvProduction, false},
	}
	for _, tc := range cases {
		cfg := config.Config{DeployEnv: tc.env}
		if got := cfg.AllowsDevHTTPRoutes(); got != tc.want {
			t.Fatalf("DeployEnv=%q AllowsDevHTTPRoutes()=%v want %v", tc.env, got, tc.want)
		}
	}
}
