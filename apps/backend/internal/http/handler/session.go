package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/http/handler/shared"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/identity/httpx"
)

type SessionHandler struct {
	shared.ProtectedHandlerBase
}

func NewSessionHandler(p httpdeps.Protected) *SessionHandler {
	return &SessionHandler{
		ProtectedHandlerBase: shared.NewProtectedHandlerBase(p),
	}
}

func (h *SessionHandler) RegisterRoutes(r chi.Router) {
	r.Get("/session", h.Get)
}

func (h *SessionHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims, err := httpx.ParseMemberToken(r, h.SessionToken)
	if err != nil {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}
	ctx, err := h.AuthzSvc.GetSessionContext(r.Context(), claims.CompanyID, claims.Subject)
	if err != nil {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}
	httpx.SetAuthzRevisionHeader(w, ctx.AuthzRevision)
	httputil.WriteJSON(w, http.StatusOK, ctx, nil)
}
