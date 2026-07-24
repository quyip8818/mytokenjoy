//go:build testhook

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/tests/testutil/pg"
)

func main() {
	baseURL := os.Getenv("DATABASE_URL")
	if baseURL == "" {
		baseURL = config.DefaultDatabaseURL
	}
	if baseURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL required")
		os.Exit(1)
	}
	if err := pg.DropOrphanTestDatabases(context.Background(), baseURL); err != nil {
		fmt.Fprintf(os.Stderr, "cleanup test databases: %v\n", err)
		os.Exit(1)
	}
}
