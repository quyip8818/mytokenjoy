package deps

import (
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
	domainpricing "github.com/tokenjoy/backend/internal/domain/pricing"
	"github.com/tokenjoy/backend/internal/identity/credentials"
	"github.com/tokenjoy/backend/internal/identity/sessiontoken"
	"github.com/tokenjoy/backend/internal/store"
)

type Public struct {
	Cfg          config.Config
	Credentials  credentials.Service
	SessionToken sessiontoken.Issuer
	SecureCookie bool
}

type Platform struct {
	Public
	Sessions       store.SessionRepository
	CompanySvc     domaincompany.Service
	BillingSvc     domainbilling.Service
	KeysSvc        domainkeys.Service
	AdminPort      adminport.Port
	PricingSvc     *domainpricing.Service
	Models         store.ModelsRepository
	SystemSettings store.SystemSettingsRepository
	PlatformQuery  store.PlatformQueryRepository
	Billing        store.BillingRepository
	Companies      store.CompanyRepository // direct repo access for register-local
	Users          store.UserRepository    // user creation for register-local
}

func (d Deps) Public() Public {
	return Public{
		Cfg:          d.Config,
		Credentials:  d.Credentials,
		SessionToken: d.SessionToken,
		SecureCookie: d.Config.SecureCookie,
	}
}

func (d Deps) Platform() Platform {
	pub := d.Public()
	return Platform{
		Public:         pub,
		Sessions:       d.Store.Session(),
		CompanySvc:     d.CompanySvc,
		BillingSvc:     d.BillingSvc,
		KeysSvc:        d.KeysSvc,
		AdminPort:      d.AdminPort,
		PricingSvc:     d.PricingSvc,
		Models:         d.Store.Models(),
		SystemSettings: d.Store.SystemSettings(),
		PlatformQuery:  d.Store.PlatformQuery(),
		Billing:        d.Store.Billing(),
		Companies:      d.Store.Company(),
		Users:          d.Store.User(),
	}
}
