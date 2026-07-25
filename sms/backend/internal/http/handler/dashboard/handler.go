package dashboard

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	dashboardsvc "sms/backend/internal/domain/dashboard"
	"sms/backend/internal/http/response"
)

type Handler struct {
	svc *dashboardsvc.Service
}

func NewHandler(svc *dashboardsvc.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/cards", h.Cards)
	r.Get("/charts", h.Charts)
	r.Get("/expiring", h.Expiring)
	r.Get("/recent-orders", h.RecentOrders)
}

func (h *Handler) Cards(w http.ResponseWriter, r *http.Request) {
	cards, err := h.svc.Cards(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "查询失败")
		return
	}
	response.JSON(w, http.StatusOK, cards)
}

func (h *Handler) Charts(w http.ResponseWriter, r *http.Request) {
	charts, err := h.svc.Charts(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "查询失败")
		return
	}
	response.JSON(w, http.StatusOK, charts)
}

func (h *Handler) Expiring(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.Expiring(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "查询失败")
		return
	}
	response.JSON(w, http.StatusOK, items)
}

func (h *Handler) RecentOrders(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.RecentOrders(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "查询失败")
		return
	}
	response.JSON(w, http.StatusOK, items)
}
