package deps

import (
	"github.com/tokenjoy/backend/internal/config"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
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
	Sessions   store.SessionRepository
	CompanySvc domaincompany.Service
	BillingSvc domainbilling.Service
	KeysSvc    domainkeys.Service
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
		Public:     pub,
		Sessions:   d.Store.Session(),
		CompanySvc: d.CompanySvc,
		BillingSvc: d.BillingSvc,
		KeysSvc:    d.KeysSvc,
	}
}
