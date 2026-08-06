package seed

import (
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/support/clock"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/snapshot"
)

func Load(cfg config.Config, clk clock.Clock) store.Snapshot {
	return snapshot.Build(cfg, clk)
}

func LoadMinimal(cfg config.Config, clk clock.Clock) store.Snapshot {
	return snapshot.BuildMinimal(cfg, clk)
}
