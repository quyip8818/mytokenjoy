package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tokenjoy/backend/internal/app"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/store/postgres"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// 1. Open connection pool (without schema/seed — those are handled below)
	pool, err := openPool(ctx, cfg)
	if err != nil {
		logger.Error("open database pool", "error", err)
		os.Exit(1)
	}

	// 2. Apply DDL schema first (setupServer needs tables to exist)
	if err := postgres.ApplySchema(ctx, pool, cfg); err != nil {
		pool.Close()
		logger.Error("apply schema", "error", err)
		os.Exit(1)
	}

	// 3. Resolve company ID (from system_settings or SaaS constant)
	companyID, err := app.ResolveCompanyID(ctx, pool, cfg)
	if err != nil {
		pool.Close()
		logger.Error("resolve company ID", "error", err)
		os.Exit(1)
	}

	// 4. If not initialized (local mode, first start) → run setup server
	if companyID == uuid.Nil {
		setupCtx, setupCancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
		defer setupCancel()

		result, err := app.RunSetupServer(setupCtx, pool, cfg, logger)
		if err != nil {
			pool.Close()
			logger.Error("setup server", "error", err)
			os.Exit(1)
		}
		companyID = result.CompanyID
	}

	pool.Close() // Close raw pool; app.New will open its own managed store

	// 5. Set runtime config — CompanyID resolved, schema already applied
	cfg.CompanyID = companyID
	cfg.StoreBootstrap.SkipSchema = true // already applied in step 2

	// 6. Normal startup: app.New (opens store → runs seed.Init with CompanyID)
	application, err := app.New(cfg, logger)
	if err != nil {
		logger.Error("create app", "error", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           application.Router,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		logger.Info("server starting", "port", cfg.Port, "deploy_env", cfg.DeployEnv, "saas", cfg.SupportSaas)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	application.Close()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown", "error", err)
	}
}

// openPool creates a raw pgxpool without any schema or seed logic.
func openPool(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	poolCfg.MaxConns = cfg.DBPoolMaxConns()
	poolCfg.MinConns = cfg.DBPoolMinConns()
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return pool, nil
}
