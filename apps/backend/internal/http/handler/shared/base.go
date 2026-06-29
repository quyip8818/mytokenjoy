package shared

import (
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/session"
)

type SessionHandlerBase struct {
	Cfg        config.Config
	SessionSvc session.Service
}

func NewSessionHandlerBase(cfg config.Config, sessionSvc session.Service) SessionHandlerBase {
	return SessionHandlerBase{Cfg: cfg, SessionSvc: sessionSvc}
}
