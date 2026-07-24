package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/billing"
	domaingateway "github.com/tokenjoy/backend/internal/domain/gateway"
	"github.com/tokenjoy/backend/internal/identity/authz"
	"github.com/tokenjoy/backend/internal/identity/credentials"
	"github.com/tokenjoy/backend/internal/identity/sessiontoken"
	"github.com/tokenjoy/backend/internal/infra/budgetcheck"
	"github.com/tokenjoy/backend/internal/infra/gatewaymetrics"
	"github.com/tokenjoy/backend/internal/store"
)

func wireIdentity(cfg config.Config, st store.Store) (authz.Service, credentials.Service, sessiontoken.Issuer, error) {
	memberToken, err := sessiontoken.NewIssuer(cfg.SessionSecret, cfg.SessionTTLSec)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("member session token: %w", err)
	}
	chargeRate := authz.ChargeRateResolver(func(ctx context.Context, companyID uuid.UUID) (string, int64, error) {
		return billing.ResolveCompanyChargeRate(ctx, st, companyID)
	})
	return authz.NewService(cfg, st, chargeRate), credentials.NewService(cfg, st), memberToken, nil
}

func wirePrecheckService(cfg config.Config, i infra) domaingateway.Prechecker {
	return domaingateway.NewPrecheckService(i.precheckCache, cfg.Clock(), budgetcheck.WrapStore(i.budgetCheck))
}

func wireGatewayService(cfg config.Config, i infra, logger *slog.Logger) (domaingateway.GatewayService, error) {
	return domaingateway.NewGatewayService(cfg, wirePrecheckService(cfg, i), i.rateLimiter, logger, gatewaymetrics.NewRecorder())
}
