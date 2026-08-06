package core

import (
	"context"
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/grants"
	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
	pkgbudget "github.com/tokenjoy/backend/internal/support/budget"
	"github.com/tokenjoy/backend/internal/support/clock"
	"github.com/tokenjoy/backend/internal/support/invitetoken"
	"github.com/tokenjoy/backend/internal/support/simulate"
)

// Store is the narrow store surface the org domain needs.
type Store interface {
	Org() store.OrgRepository
	User() store.UserRepository
	Company() store.CompanyRepository
	Invite() store.InviteRepository
	SchedulerLock() store.SchedulerLockRepository
	TenantBackgroundState() store.TenantBackgroundStateRepository
	RiverJob() store.RiverJobRepository
	WithTx(ctx context.Context, fn func(store.Store) error) error
}

// DirectSender delivers messages (SMS/email) without recipient resolution.
type DirectSender interface {
	SendDirect(ctx context.Context, channel string, address string, msg domainnotification.RenderedMessage) error
}

type Deps struct {
	Cfg          config.Config
	Clock        clock.Clock
	Store        Store
	Factory      types.DataSourceFactory
	Notifier     types.Notifier
	Sender       DirectSender
	InviteIssuer *invitetoken.Issuer // nil if INVITE_SECRET not configured
	Delayer      simulate.Delayer
	Logger       *slog.Logger
	Grants       grants.Normalizer
	cryptoKey    []byte
}

func NewDeps(
	cfg config.Config,
	st Store,
	factory types.DataSourceFactory,
	notifier types.Notifier,
	sender DirectSender,
	delayer simulate.Delayer,
	logger *slog.Logger,
	grants grants.Normalizer,
	clk clock.Clock,
) *Deps {
	if logger == nil {
		logger = slog.Default()
	}

	// Build invite token issuer if secret is configured.
	var issuer *invitetoken.Issuer
	if keys := cfg.InviteSecretKeys(); len(keys) > 0 {
		iss, err := invitetoken.NewIssuer(keys...)
		if err == nil {
			issuer = iss
		} else {
			logger.Warn("invitetoken: failed to create issuer", "error", err)
		}
	}

	return &Deps{
		Cfg:          cfg,
		Clock:        clock.OrDefault(clk),
		Store:        st,
		Factory:      factory,
		Notifier:     notifier,
		Sender:       sender,
		InviteIssuer: issuer,
		Delayer:      delayer,
		Logger:       logger,
		Grants:       grants,
	}
}

func (d *Deps) BudgetPeriod() string {
	return pkgbudget.PeriodMonthly
}
