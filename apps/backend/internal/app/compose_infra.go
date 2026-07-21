package app

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/tokenjoy/backend/internal/adapter"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	domaingateway "github.com/tokenjoy/backend/internal/domain/gateway"
	"github.com/tokenjoy/backend/internal/domain/newapisync"
	"github.com/tokenjoy/backend/internal/domain/newapisync/policy"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/identity/verifycode"
	"github.com/tokenjoy/backend/internal/infra/budgetcheck"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/infra/ratelimit"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/common"
	pkgrl "github.com/tokenjoy/backend/internal/pkg/ratelimit"
	"github.com/tokenjoy/backend/internal/store"
)

type infra struct {
	store           store.Store
	adminPort       adminport.Port
	newAPISync      newapisync.Lifecycle
	channelPolicy   policy.ChannelPolicy
	companyGate     *domaincompany.Gate
	notifier        types.Notifier
	notificationSvc *notification.Service
	delayer         common.Delayer
	enqueuer        jobs.Enqueuer
	budgetCheck     budgetcheck.Store
	rateLimiter     pkgrl.Limiter
	precheckCache   *domaingateway.PrecheckCache
	verifyCodeSvc   *verifycode.Service
}

func buildInfraWithStore(cfg config.Config, logger *slog.Logger, st store.Store, enqueuer jobs.Enqueuer, adminPortOverride adminport.Port) (infra, error) {
	var adminPort adminport.Port
	if adminPortOverride != nil {
		adminPort = adminPortOverride
	} else {
		port, err := buildAdminPort(context.Background(), cfg, logger)
		if err != nil {
			return infra{}, err
		}
		adminPort = port
	}
	if enqueuer == nil {
		enqueuer = jobs.NoopEnqueuer{}
	}
	channelPolicy := policy.NewChannelPolicy(cfg)

	notifySvc := notification.NewService(cfg, st, logger)

	return infra{
		store:           st,
		adminPort:       adminPort,
		channelPolicy:   channelPolicy,
		companyGate:     domaincompany.NewGate(cfg),
		newAPISync:      newapisync.New(cfg, st, adminPort, channelPolicy, adapter.NewNewAPISyncEnqueuer(enqueuer)),
		notifier:        notifySvc,
		notificationSvc: notifySvc,
		delayer:         common.NewDelayer(cfg.SimulateDelay),
		enqueuer:        enqueuer,
		budgetCheck:     budgetcheck.Open(context.Background(), cfg, logger),
		rateLimiter:     ratelimit.Open(context.Background(), cfg.RedisURL, logger),
		precheckCache:   domaingateway.NewPrecheckCache(st.GatewayPrecheck()),
		verifyCodeSvc:   buildVerifyCodeService(cfg, notifySvc, logger),
	}, nil
}

func buildVerifyCodeService(cfg config.Config, notifier verifycode.Notifier, logger *slog.Logger) *verifycode.Service {
	if !cfg.SupportSaas {
		return nil
	}
	svc, err := verifycode.NewService(verifycode.Config{
		RedisURL:   cfg.RedisURL,
		SkipVerify: cfg.SkipVerifyCode,
	}, notifier, logger)
	if err != nil {
		logger.Error("verifycode: service init failed", "error", err)
		return nil
	}
	return svc
}

// buildAdminPort creates the NewAPI admin port by reading the admin token
// directly from NewAPI's database (same Postgres instance, "newapi" database).
func buildAdminPort(ctx context.Context, cfg config.Config, logger *slog.Logger) (adminport.Port, error) {
	if !cfg.NewAPIEnabled || strings.TrimSpace(cfg.NewAPIBaseURL) == "" {
		return nil, nil
	}
	newAPIDSN := cfg.NewAPIDatabaseURL
	if newAPIDSN == "" {
		derived, err := newapi.DeriveNewAPIDSN(cfg.DatabaseURL)
		if err != nil {
			return nil, fmt.Errorf("derive NewAPI DSN: %w", err)
		}
		newAPIDSN = derived
	}
	tokenStore := newapi.NewTokenStore(newAPIDSN, cfg.NewAPIAdminUserID)
	token, err := tokenStore.FetchToken(ctx)
	if err != nil {
		return nil, newapi.FormatError(err)
	}
	logger.Info("newapi admin token loaded from database")
	client := newapi.NewClient(cfg.NewAPIBaseURL, token, cfg.NewAPIAdminUserID)
	return newapi.NewSelfHealingPort(client, tokenStore), nil
}
