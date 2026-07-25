package app

import (
	"context"
	"log/slog"
	"net/http"

	"sms/backend/internal/config"
	authsvc "sms/backend/internal/domain/auth"
	contractsvc "sms/backend/internal/domain/contract"
	dashboardsvc "sms/backend/internal/domain/dashboard"
	evalsvc "sms/backend/internal/domain/evaluation"
	modelsvc "sms/backend/internal/domain/model"
	newapisyncsvc "sms/backend/internal/domain/newapisync"
	ordersvc "sms/backend/internal/domain/order"
	suppliersvc "sms/backend/internal/domain/supplier"
	usersvc "sms/backend/internal/domain/user"
	httpapi "sms/backend/internal/http"
	"sms/backend/internal/http/deps"
	newapiclient "sms/backend/internal/integration/newapi"
	"sms/backend/internal/store"
	"sms/backend/internal/store/postgres"
)

type App struct {
	Router     http.Handler
	pool       interface{ Close() }
	newapiPool interface{ Close() } // nil when disabled
}

func New(cfg config.Config, logger *slog.Logger) (*App, error) {
	pool, err := store.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	pg := postgres.New(pool)

	authService := authsvc.NewService(pg, cfg.JWTSecret, cfg.AccessTokenTTL(), cfg.RefreshTokenTTL())
	supplierService := suppliersvc.NewService(pg)
	modelService := modelsvc.NewService(pg)
	contractService := contractsvc.NewService(pg, cfg.UploadDir)
	orderService := ordersvc.NewService(pg)
	evalService := evalsvc.NewService(pg)
	dashboardService := dashboardsvc.NewService(pg)
	userService := usersvc.NewService(pg)

	var syncService *newapisyncsvc.Service
	var newapiPool interface{ Close() }

	if cfg.NewAPIEnabled() {
		naPool, err := store.NewPool(context.Background(), cfg.NewAPIDBURL())
		if err != nil {
			logger.Warn("failed to connect newapi DB, sync disabled", "error", err)
		} else {
			newapiPool = naPool
			tokenStore := newapiclient.NewTokenStore(naPool, cfg.NewAPIAdminUserID)
			client := newapiclient.NewClient(cfg.NewAPIBaseURL, tokenStore)
			syncService = newapisyncsvc.NewService(client, modelService, logger)
			modelService.SetSyncer(syncService)
			logger.Info("newapi sync enabled", "baseURL", cfg.NewAPIBaseURL)
		}
	}

	d := deps.Deps{
		Config:     cfg,
		Logger:     logger,
		Auth:       authService,
		Supplier:   supplierService,
		Model:      modelService,
		Contract:   contractService,
		Order:      orderService,
		Eval:       evalService,
		Dashboard:  dashboardService,
		User:       userService,
		NewAPISync: syncService,
	}

	router := httpapi.NewRouter(d)

	return &App{Router: router, pool: pool, newapiPool: newapiPool}, nil
}

func (a *App) Close() {
	a.pool.Close()
	if a.newapiPool != nil {
		a.newapiPool.Close()
	}
}
