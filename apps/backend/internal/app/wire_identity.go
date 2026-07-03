package app

import (
	"fmt"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/identity/authz"
	"github.com/tokenjoy/backend/internal/identity/credentials"
	"github.com/tokenjoy/backend/internal/identity/sessiontoken"
	"github.com/tokenjoy/backend/internal/store"
)

func wireIdentity(cfg config.Config, st store.Store) (authz.Service, credentials.Service, sessiontoken.Issuer, sessiontoken.Issuer, error) {
	memberToken, err := sessiontoken.NewIssuer(cfg.SessionSecret, cfg.SessionTTLSec)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("member session token: %w", err)
	}
	platformToken, err := sessiontoken.NewIssuer(cfg.ResolvedPlatformSessionSecret(), cfg.SessionTTLSec)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("platform session token: %w", err)
	}
	return authz.NewService(cfg, st), credentials.NewService(cfg, st), memberToken, platformToken, nil
}
