package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tokenjoy/backend/internal/domain/session"
	"github.com/tokenjoy/backend/internal/http/response"
	"github.com/tokenjoy/backend/internal/pkg/sessionutil"
)

type SessionHandler struct {
	service session.Service
}

func NewSessionHandler(service session.Service) *SessionHandler {
	return &SessionHandler{service: service}
}

func (h *SessionHandler) RegisterRoutes(r chi.Router) {
	r.Get("/session", h.Get)
}

func (h *SessionHandler) Get(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if _, hasMemberID := query["memberId"]; hasMemberID {
		memberID := query.Get("memberId")
		if memberID == "" {
			response.Error(w, http.StatusBadRequest, "memberId is required")
			return
		}
		h.respondSession(w, memberID)
		return
	}

	memberID := sessionutil.ResolveMemberID(r)
	if memberID == "" {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	h.respondSession(w, memberID)
}

func (h *SessionHandler) respondSession(w http.ResponseWriter, memberID string) {
	ctx, err := h.service.GetByMemberID(memberID)
	if err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, ctx)
}
