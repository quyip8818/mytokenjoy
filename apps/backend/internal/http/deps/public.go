package deps

import (
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/identity/credentials"
	"github.com/tokenjoy/backend/internal/domain/identity/sessiontoken"
	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
	domainplatform "github.com/tokenjoy/backend/internal/domain/platform"
	domainredemption "github.com/tokenjoy/backend/internal/domain/redemption"
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
	PlatformSvc    domainplatform.Service
	AdminPort      adminport.Port
	RedemptionSvc  domainredemption.Service
	Models         store.ModelsRepository
	ModelDiscount  store.ModelDiscountRepository
	SystemSettings store.SystemSettingsRepository
	SyncVersions   store.SyncVersionRepository
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
		PlatformSvc:    domainplatform.NewService(d.Store, d.AdminPort, d.Config.TokenJoyCompanyID),
		AdminPort:      d.AdminPort,
		RedemptionSvc:  domainredemption.NewService(d.Store),
		Models:         d.Store.Models(),
		ModelDiscount:  d.Store.ModelDiscount(),
		SystemSettings: d.Store.SystemSettings(),
		SyncVersions:   d.Store.SyncVersions(),
		PlatformQuery:  d.Store.PlatformQuery(),
		Billing:        d.Store.Billing(),
		Companies:      d.Store.Company(),
		Users:          d.Store.User(),
	}
}
