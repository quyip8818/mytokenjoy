package dev

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/newapisync/devapi"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/http/handler/shared"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/infra/permission"
)

// Handler serves /api/dev/* — registered only when config.AllowsDevHTTPRoutes.
type Handler struct {
	shared.ProtectedHandlerBase
	localCompanyID   int64
	bearerResolver   devapi.BearerResolver
	readinessChecker devapi.ReadinessChecker
}

func NewHandler(
	p httpdeps.Protected,
	localCompanyID int64,
	bearerResolver devapi.BearerResolver,
	readinessChecker devapi.ReadinessChecker,
) *Handler {
	return &Handler{
		ProtectedHandlerBase: shared.NewProtectedHandlerBase(p),
		localCompanyID:       localCompanyID,
		bearerResolver:       bearerResolver,
		readinessChecker:     readinessChecker,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/readiness", h.Readiness)
	read := httpmiddleware.ReadRoutes(r, h.Protected, permission.KeysAdmin)
	read.Get("/platform-keys/{id}/bearer", h.PlatformKeyBearer)
}

type readinessResponse struct {
	Ready                 bool     `json:"ready"`
	UnreadyPlatformKeyIDs []string `json:"unready_platform_key_ids,omitempty"`
}

func (h *Handler) Readiness(w http.ResponseWriter, r *http.Request) {
	if h.readinessChecker == nil {
		httputil.WriteJSON(w, http.StatusServiceUnavailable, readinessResponse{Ready: false}, nil)
		return
	}
	ctx := company.DefaultContext(h.localCompanyID)
	unready, err := h.readinessChecker.UnreadyPlatformKeyIDs(ctx)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	if len(unready) > 0 {
		httputil.WriteJSON(w, http.StatusServiceUnavailable, readinessResponse{
			Ready:                 false,
			UnreadyPlatformKeyIDs: unready,
		}, nil)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, readinessResponse{Ready: true}, nil)
}

type platformKeyBearerResponse struct {
	Bearer string `json:"bearer"`
}

func (h *Handler) PlatformKeyBearer(w http.ResponseWriter, r *http.Request) {
	ctx := company.DefaultContext(h.localCompanyID)
	bearer, err := h.bearerResolver.ResolvePlatformKeyBearer(ctx, chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, platformKeyBearerResponse{Bearer: bearer}, nil)
}
