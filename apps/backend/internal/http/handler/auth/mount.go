package auth

import (
	"github.com/go-chi/chi/v5"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/domain/identity/registertoken"
	"github.com/tokenjoy/backend/internal/support/invitetoken"
)

// Mount registers the auth handler on the given router.
func Mount(r chi.Router, d httpdeps.Deps) {
	regTokenIssuer := registertoken.NewIssuer(d.SessionToken.Secret())

	// Build invite token issuer (nil if INVITE_SECRET not configured).
	var invTokenIssuer *invitetoken.Issuer
	if keys := d.Config.InviteSecretKeys(); len(keys) > 0 {
		iss, err := invitetoken.NewIssuer(keys...)
		if err == nil {
			invTokenIssuer = iss
		}
	}

	h := NewHandler(d.Public(), d.CompanySvc, d.Users(), d.Sessions(), d.Invites(), d.Org(), d.Company(), d.VerifyCodeSvc, regTokenIssuer, invTokenIssuer)
	h.RegisterRoutes(r)
}
