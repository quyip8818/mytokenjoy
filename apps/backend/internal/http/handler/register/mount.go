package register

import (
	"github.com/go-chi/chi/v5"
	"github.com/tokenjoy/backend/internal/domain/identity/registertoken"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
)

// Mount registers the register handler on the given router (SaaS mode only).
func Mount(r chi.Router, d httpdeps.Deps) {
	regTokenIssuer := registertoken.NewIssuer(d.SessionToken.Secret())
	h := NewHandler(
		d.CompanySvc, d.Users(), d.Sessions(), d.VerifyCodeSvc,
		regTokenIssuer, d.SessionToken,
		d.Config.SecureCookie, d.Config.RegistrationEnabled,
		d.Config.SessionTTLSec, d.Config.RefreshTokenTTLSec,
	)
	h.RegisterRoutes(r)
}
