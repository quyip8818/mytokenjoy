package app

import (
	"github.com/tokenjoy/backend/internal/domain/session"
	"github.com/tokenjoy/backend/internal/store"
)

func wireSession(st store.Store) session.Service {
	return session.NewService(st)
}
