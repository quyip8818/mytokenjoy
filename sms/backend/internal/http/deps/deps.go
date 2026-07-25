package deps

import (
	"log/slog"

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
)

type Deps struct {
	Config     config.Config
	Logger     *slog.Logger
	Auth       *authsvc.Service
	Supplier   *suppliersvc.Service
	Model      *modelsvc.Service
	Contract   *contractsvc.Service
	Order      *ordersvc.Service
	Eval       *evalsvc.Service
	Dashboard  *dashboardsvc.Service
	User       *usersvc.Service
	NewAPISync *newapisyncsvc.Service // nil when NEWAPI_BASE_URL is empty
}
