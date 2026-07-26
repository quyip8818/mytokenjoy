package oauth

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	oauthsvc "sms/backend/internal/domain/oauth"
)

type Handler struct {
	service *oauthsvc.Service
}

func NewHandler(service *oauthsvc.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/token", h.Token)
}

func (h *Handler) Token(w http.ResponseWriter, r *http.Request) {
	var body struct {
		GrantType    string `json:"grant_type"`
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"message":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if body.GrantType != "client_credentials" {
		http.Error(w, `{"message":"unsupported grant_type"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.service.IssueToken(r.Context(), body.ClientID, body.ClientSecret)
	if err != nil {
		http.Error(w, `{"message":"invalid client credentials"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
