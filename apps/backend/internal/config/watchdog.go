package config

import "time"

func (c Config) WatchdogInterval() time.Duration {
	if c.WatchdogIntervalSec <= 0 {
		return 7 * 24 * time.Hour
	}
	return time.Duration(c.WatchdogIntervalSec) * time.Second
}

func (c Config) WatchdogBulkBatchSize() int {
	if c.WatchdogBulkBatchSizeEnv <= 0 {
		return 200
	}
	return c.WatchdogBulkBatchSizeEnv
}

func (c Config) WatchdogStartupDelay() time.Duration {
	if c.WatchdogStartupDelaySec <= 0 {
		return 5 * time.Second
	}
	return time.Duration(c.WatchdogStartupDelaySec) * time.Second
}
