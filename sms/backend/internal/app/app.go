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
	ordersvc "sms/backend/internal/domain/order"
	suppliersvc "sms/backend/internal/domain/supplier"
	usersvc "sms/backend/internal/domain/user"
	httpapi "sms/backend/internal/http"
	"sms/backend/internal/http/deps"
	"sms/backend/internal/store"
	"sms/backend/internal/store/postgres"
)

type App struct {
	Router http.Handler
	pool   interface{ Close() }
}

func New(cfg config.Config, logger *slog.Logger) (*App, error) {
	pool, err := store.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	pg := postgres.New(pool)

	authService := authsvc.NewService(pg, cfg.JWTSecret, cfg.AccessTokenTTL(), cfg.RefreshTokenTTL())
	supplierService := suppliersvc.NewService(pg)
	contractService := contractsvc.NewService(pg, cfg.UploadDir)
	orderService := ordersvc.NewService(pg)
	evalService := evalsvc.NewService(pg)
	dashboardService := dashboardsvc.NewService(pg)
	userService := usersvc.NewService(pg)

	d := deps.Deps{
		Config:    cfg,
		Logger:    logger,
		Auth:      authService,
		Supplier:  supplierService,
		Contract:  contractService,
		Order:     orderService,
		Eval:      evalService,
		Dashboard: dashboardService,
		User:      userService,
	}

	router := httpapi.NewRouter(d)

	return &App{Router: router, pool: pool}, nil
}

func (a *App) Close() {
	a.pool.Close()
}
