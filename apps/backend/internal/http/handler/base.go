package handler

import (
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/session"
)

type sessionHandlerBase struct {
	cfg        config.Config
	sessionSvc session.Service
}

func newSessionHandlerBase(cfg config.Config, sessionSvc session.Service) sessionHandlerBase {
	return sessionHandlerBase{cfg: cfg, sessionSvc: sessionSvc}
}
