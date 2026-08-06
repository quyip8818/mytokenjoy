package deps

import (
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/identity/registertoken"
	"github.com/tokenjoy/backend/internal/domain/identity/verifycode"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/support/invitetoken"
)

// Auth holds the narrowed dependencies for the auth handler (login/logout/invite flow).
type Auth struct {
	Public
	CompanySvc    domaincompany.Service
	Users         store.UserRepository
	Sessions      store.SessionRepository
	Invites       store.InviteRepository
	OrgRepo       store.OrgRepository
	Companies     store.CompanyRepository
	VerifyCode    *verifycode.Service
	RegisterToken *registertoken.Issuer
	InviteToken   *invitetoken.Issuer
}

func (d Deps) Auth() Auth {
	regTokenIssuer := registertoken.NewIssuer(d.SessionToken.Secret())

	var invTokenIssuer *invitetoken.Issuer
	if keys := d.Config.InviteSecretKeys(); len(keys) > 0 {
		iss, err := invitetoken.NewIssuer(keys...)
		if err == nil {
			invTokenIssuer = iss
		}
	}

	return Auth{
		Public:        d.Public(),
		CompanySvc:    d.CompanySvc,
		Users:         d.Store.User(),
		Sessions:      d.Store.Session(),
		Invites:       d.Store.Invite(),
		OrgRepo:       d.Store.Org(),
		Companies:     d.Store.Company(),
		VerifyCode:    d.VerifyCodeSvc,
		RegisterToken: regTokenIssuer,
		InviteToken:   invTokenIssuer,
	}
}
