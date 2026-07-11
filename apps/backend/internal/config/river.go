package config

import "time"

const (
	RiverQueueCritical = "critical"
	RiverQueueDefault  = "default"
	RiverQueueLow      = "low"
)

// RiverConfig holds River worker settings (embedded in Config for env parsing).
type RiverConfig struct {
	RiverEnabled       bool `env:"RIVER_ENABLED" envDefault:"true"`
	RiverMaxWorkersEnv int  `env:"RIVER_MAX_WORKERS" envDefault:"20"`
}

func (c Config) RiverMaxWorkers() int {
	if c.RiverMaxWorkersEnv <= 0 {
		return 20
	}
	return c.RiverMaxWorkersEnv
}

func (c Config) WorkerBudgetReconcileInterval() time.Duration {
	return 30 * time.Minute
}

func (c Config) WorkerDashboardProjectInterval() time.Duration {
	return time.Hour
}

func (c Config) WorkerDashboardReconcileInterval() time.Duration {
	return 24 * time.Hour
}
