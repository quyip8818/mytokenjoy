package deps

import (
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/identity/authz"
	"github.com/tokenjoy/backend/internal/domain/identity/sessiontoken"
)

type Protected struct {
	Cfg          config.Config
	AuthzSvc     authz.Service
	SessionToken sessiontoken.Issuer
}

func (d Deps) Protected() Protected {
	return Protected{
		Cfg:          d.Config,
		AuthzSvc:     d.AuthzSvc,
		SessionToken: d.SessionToken,
	}
}
