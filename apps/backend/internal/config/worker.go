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
