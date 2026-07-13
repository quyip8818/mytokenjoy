package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/config"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}
	if err := app.RunDevBootstrap(context.Background(), cfg, logger); err != nil {
		logger.Error("dev bootstrap failed", "error", err)
		os.Exit(1)
	}
	logger.Info("dev bootstrap complete")
}
