package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tokenjoy/backend/internal/config"
)

var monthlyPartitionedTables = []string{
	"operation_logs",
	"usage_ledger",
	"usage_buckets",
}

const (
	partitionStartYear  = 2024
	partitionStartMonth = 1
	partitionEndYear    = 2032
	partitionEndMonth   = 12
)

func applyMonthlyPartitions(ctx context.Context, pool dbQuerier, cfg config.Config) error {
	start, endBound := partitionBounds(cfg)

	for _, table := range monthlyPartitionedTables {
		var ddl strings.Builder
		for cursor := start; cursor.Before(endBound); cursor = cursor.AddDate(0, 1, 0) {
			next := cursor.AddDate(0, 1, 0)
			partition := fmt.Sprintf("%s_%04d_%02d", table, cursor.Year(), int(cursor.Month()))
			fmt.Fprintf(
				&ddl,
				"CREATE TABLE IF NOT EXISTS %s PARTITION OF %s FOR VALUES FROM ('%s') TO ('%s');\n",
				partition,
				table,
				cursor.Format("2006-01-02"),
				next.Format("2006-01-02"),
			)
		}
		if _, err := pool.Exec(ctx, ddl.String()); err != nil {
			return fmt.Errorf("create %s partitions: %w", table, err)
		}
	}
	return nil
}

func partitionBounds(cfg config.Config) (time.Time, time.Time) {
	if cfg.StoreBootstrap.TestPartitionMonths > 0 {
		start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		endBound := start.AddDate(0, cfg.StoreBootstrap.TestPartitionMonths, 0)
		return start, endBound
	}
	start := time.Date(partitionStartYear, partitionStartMonth, 1, 0, 0, 0, 0, time.UTC)
	endBound := time.Date(partitionEndYear, partitionEndMonth, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0)
	return start, endBound
}
