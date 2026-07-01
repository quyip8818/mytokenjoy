package authhandler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/pkg/common"
)

type Handler struct {
	companySvc domaincompany.Service
}

func NewHandler(companySvc domaincompany.Service) *Handler {
	return &Handler{companySvc: companySvc}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/auth/accept-invite", h.AcceptInvite)
}

type acceptInviteBody struct {
	Token    string `json:"token"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (h *Handler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	var body acceptInviteBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	member, err := h.companySvc.AcceptInvite(r.Context(), domaincompany.AcceptInviteRequest{
		Token: body.Token, Name: body.Name, Password: body.Password,
	})
	if err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, nil, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: common.SessionCookie, Value: member.ID, Path: "/",
		HttpOnly: true, SameSite: http.SameSiteLaxMode,
	})
	httputil.WriteJSON(w, http.StatusOK, member, nil)
}
