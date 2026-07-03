package deps

import (
	"github.com/tokenjoy/backend/internal/config"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
	"github.com/tokenjoy/backend/internal/identity/credentials"
	"github.com/tokenjoy/backend/internal/identity/sessiontoken"
)

type Public struct {
	Cfg                  config.Config
	Credentials          credentials.Service
	SessionToken         sessiontoken.Issuer
	PlatformSessionToken sessiontoken.Issuer
	SecureCookie         bool
}

type Platform struct {
	Public
	CompanySvc domaincompany.Service
	BillingSvc domainbilling.Service
	KeysSvc    domainkeys.Service
}

func (d Deps) Public() Public {
	return Public{
		Cfg:                  d.Config,
		Credentials:          d.Credentials,
		SessionToken:         d.SessionToken,
		PlatformSessionToken: d.PlatformSessionToken,
		SecureCookie:         d.Config.IsProdProfile(),
	}
}

func (d Deps) Platform() Platform {
	pub := d.Public()
	return Platform{
		Public:     pub,
		CompanySvc: d.CompanySvc,
		BillingSvc: d.BillingSvc,
		KeysSvc:    d.KeysSvc,
	}
}
