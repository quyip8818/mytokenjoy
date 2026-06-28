package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/domain/session"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/pkg/sessionutil"
)

type SessionHandler struct {
	cfg     config.Config
	service session.Service
}

func NewSessionHandler(cfg config.Config, service session.Service) *SessionHandler {
	return &SessionHandler{cfg: cfg, service: service}
}

func (h *SessionHandler) RegisterRoutes(r chi.Router) {
	r.Get("/session", h.Get)
}

func (h *SessionHandler) Get(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if _, hasMemberID := query["memberId"]; hasMemberID && h.cfg.IsDemoProfile() {
		memberID := query.Get("memberId")
		if memberID == "" {
			httputil.WriteError(w, domain.BadRequest("memberId is required"))
			return
		}
		h.respondSession(w, memberID)
		return
	}

	memberID := sessionutil.ResolveMemberID(r, h.cfg.IsDemoProfile())
	if memberID == "" {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}
	h.respondSession(w, memberID)
}

func (h *SessionHandler) respondSession(w http.ResponseWriter, memberID string) {
	ctx, err := h.service.GetByMemberID(memberID)
	httputil.WriteJSON(w, http.StatusOK, ctx, err)
}
