package config

import "time"

func (c Config) WorkerPollInterval() time.Duration {
	if c.WorkerPollIntervalSec <= 0 {
		return 5 * time.Second
	}
	return time.Duration(c.WorkerPollIntervalSec) * time.Second
}

func (c Config) WorkerOrgSyncInterval() time.Duration {
	if c.WorkerOrgSyncIntervalSec <= 0 {
		return time.Minute
	}
	return time.Duration(c.WorkerOrgSyncIntervalSec) * time.Second
}

func (c Config) IngestReconcileInterval() time.Duration {
	if c.IngestReconcileIntervalSec <= 0 {
		return 300 * time.Second
	}
	return time.Duration(c.IngestReconcileIntervalSec) * time.Second
}

func (c Config) ReconcileBatchSize() int {
	if c.IngestReconcileBatchSize <= 0 {
		return 500
	}
	return c.IngestReconcileBatchSize
}

func (c Config) ReconcileMaxRounds() int {
	if c.IngestReconcileMaxRounds <= 0 {
		return 10
	}
	return c.IngestReconcileMaxRounds
}

func (c Config) JobBatchSize() int {
	if c.IngestJobBatchSize <= 0 {
		return 20
	}
	return c.IngestJobBatchSize
}
