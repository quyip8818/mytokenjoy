package seed

import (
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/seed/snapshot"
)

func Load(cfg config.Config) store.Snapshot {
	return snapshot.Build(cfg)
}

func LoadMinimal(cfg config.Config) store.Snapshot {
	return snapshot.BuildMinimal(cfg)
}
