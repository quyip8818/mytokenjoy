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
	"github.com/tokenjoy/backend/internal/identity/sms"
	"github.com/tokenjoy/backend/internal/infra/budgetcheck"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/infra/ratelimit"
	"github.com/tokenjoy/backend/internal/integration/newapi"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/store"
)

type infra struct {
	store           store.Store
	adminPort       adminport.Port
	newAPISync      newapisync.Lifecycle
	channelPolicy   policy.ChannelPolicy
	wallet          domaincompany.WalletService
	companyGate     *domaincompany.Gate
	notifier        types.Notifier
	notificationSvc *notification.Service
	delayer         common.Delayer
	enqueuer        jobs.Enqueuer
	budgetCheck     budgetcheck.Store
	rateLimiter     ratelimit.Limiter
	precheckCache   *domaingateway.PrecheckCache
	smsSvc          *sms.Service
}

func buildInfraWithStore(cfg config.Config, logger *slog.Logger, st store.Store, enqueuer jobs.Enqueuer, adminClientOverride newapi.AdminClient) (infra, error) {
	var adminClient newapi.AdminClient
	if adminClientOverride != nil {
		adminClient = adminClientOverride
	} else {
		adminClient = newapi.NewClient(cfg.NewAPIBaseURL, cfg.NewAPIAdminToken, cfg.NewAPIAdminUserID)
	}
	if enqueuer == nil {
		enqueuer = jobs.NoopEnqueuer{}
	}
	channelPolicy := policy.NewChannelPolicy(cfg)
	adminPort := newapi.NewAdminPortAdapter(adminClient)
	wallet := domaincompany.NewWalletService(cfg, adminPort)

	notifySvc := notification.NewService(cfg, st, logger)

	return infra{
		store:           st,
		adminPort:       adminPort,
		channelPolicy:   channelPolicy,
		wallet:          wallet,
		companyGate:     domaincompany.NewGate(cfg),
		newAPISync:      newapisync.New(cfg, st, adminPort, channelPolicy, adapter.NewNewAPISyncEnqueuer(enqueuer)),
		notifier:        notifySvc,
		notificationSvc: notifySvc,
		delayer:         common.NewDelayer(cfg.SimulateDelay),
		enqueuer:        enqueuer,
		budgetCheck:     budgetcheck.Open(context.Background(), cfg, logger),
		rateLimiter:     ratelimit.Open(context.Background(), cfg.RedisURL, logger),
		precheckCache:   domaingateway.NewPrecheckCache(st.GatewayPrecheck()),
		smsSvc:          buildSMSService(cfg, logger),
	}, nil
}


func buildSMSService(cfg config.Config, logger *slog.Logger) *sms.Service {
	if !cfg.SupportSaas {
		return nil
	}
	sender := sms.NewAliyunSender(sms.AliyunConfig{
		AccessKeyID:     cfg.AliyunSMSAccessKeyID,
		AccessKeySecret: cfg.AliyunSMSAccessKeySecret,
		SignName:        cfg.AliyunSMSSignName,
		TemplateCode:    cfg.AliyunSMSTemplateCode,
		Endpoint:        cfg.AliyunSMSEndpoint,
	}, logger)
	svc, err := sms.NewService(cfg.RedisURL, sender, logger)
	if err != nil {
		logger.Error("sms: service init failed", "error", err)
		return nil
	}
	return svc
}
