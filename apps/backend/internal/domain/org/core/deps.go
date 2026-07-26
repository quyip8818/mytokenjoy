package core

import (
	"context"
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/grants"
	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/integration/datasource"
	pkgbudget "github.com/tokenjoy/backend/internal/pkg/budget"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/invitetoken"
	"github.com/tokenjoy/backend/internal/store"
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
	Store        Store
	Factory      datasource.Factory
	Notifier     types.Notifier
	Sender       DirectSender
	InviteIssuer *invitetoken.Issuer // nil if INVITE_SECRET not configured
	Delayer      common.Delayer
	Logger       *slog.Logger
	Grants       grants.Normalizer
	cryptoKey    []byte
}

func NewDeps(
	cfg config.Config,
	st Store,
	factory datasource.Factory,
	notifier types.Notifier,
	sender DirectSender,
	delayer common.Delayer,
	logger *slog.Logger,
	grants grants.Normalizer,
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
