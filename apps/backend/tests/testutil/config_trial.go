//go:build testhook

package testutil

import "github.com/tokenjoy/backend/internal/config"

func WithTrialMemberLimit(limit int) ConfigOption {
	return func(cfg *config.Config) {
		cfg.TrialMemberLimit = limit
	}
}
