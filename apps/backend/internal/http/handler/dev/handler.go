package dev

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tokenjoy/backend/internal/domain/newapisync"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/http/handler/shared"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/infra/permission"
)

// Handler serves /api/dev/* — registered only when config.AllowsDevHTTPRoutes.
type Handler struct {
	shared.ProtectedHandlerBase
	bearerResolver newapisync.DevBearerResolver
}

func NewHandler(p httpdeps.Protected, bearerResolver newapisync.DevBearerResolver) *Handler {
	return &Handler{
		ProtectedHandlerBase: shared.NewProtectedHandlerBase(p),
		bearerResolver:       bearerResolver,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	read := httpmiddleware.ReadRoutes(r, h.Protected, permission.KeysAdmin)
	read.Get("/platform-keys/{id}/bearer", h.PlatformKeyBearer)
}

type platformKeyBearerResponse struct {
	Bearer string `json:"bearer"`
}

func (h *Handler) PlatformKeyBearer(w http.ResponseWriter, r *http.Request) {
	bearer, err := h.bearerResolver.ResolvePlatformKeyBearer(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, platformKeyBearerResponse{Bearer: bearer}, nil)
}
