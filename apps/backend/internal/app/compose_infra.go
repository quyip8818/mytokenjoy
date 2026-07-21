package app

import (
	"context"
	"log/slog"

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
		adminPort = newapi.NewPort(cfg.NewAPIBaseURL, cfg.NewAPIAdminToken, cfg.NewAPIAdminUserID)
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
