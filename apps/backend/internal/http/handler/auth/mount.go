package auth

import (
	"github.com/go-chi/chi/v5"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/identity/registertoken"
)

// Mount registers the auth handler on the given router.
func Mount(r chi.Router, d httpdeps.Deps) {
	regTokenIssuer := registertoken.NewIssuer(d.SessionToken.Secret())
	h := NewHandler(d.Public(), d.CompanySvc, d.Users(), d.Sessions(), d.Invites(), d.VerifyCodeSvc, regTokenIssuer)
	h.RegisterRoutes(r)
}
