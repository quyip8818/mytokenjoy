package notification

import (
	"github.com/go-chi/chi/v5"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/http/handler/shared"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/store"
)

// Handler serves notification-related HTTP endpoints.
type Handler struct {
	shared.ProtectedHandlerBase
	store     store.Store
	notifySvc *notification.Service
}

func NewHandler(p httpdeps.Protected, st store.Store, notifySvc *notification.Service) *Handler {
	return &Handler{
		ProtectedHandlerBase: shared.NewProtectedHandlerBase(p),
		store:                st,
		notifySvc:            notifySvc,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	session := httpmiddleware.SessionRoutes(r, h.Protected)
	session.Get("/", h.List)
	session.Get("/unread-count", h.UnreadCount)
	session.Patch("/{id}/read", h.MarkRead)
	session.Post("/read-all", h.MarkAllRead)
	session.Get("/capabilities", h.Capabilities)
	session.Get("/stream", h.Stream)
	session.Get("/preferences", h.GetPreferences)
	session.Put("/preferences", h.UpdatePreferences)
	session.Post("/preferences/reset", h.ResetPreferences)

	// Admin routes (require audit:read permission)
	admin := httpmiddleware.ReadRoutes(r, h.Protected, "audit:read")
	admin.Get("/admin/log", h.AdminLog)
	admin.Get("/admin/stats", h.AdminStats)
	admin.Post("/admin/test", h.AdminTestSend)
}
