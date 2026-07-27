package model

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	modelsvc "sms/backend/internal/domain/model"
	"sms/backend/internal/http/helpers"
	httpmiddleware "sms/backend/internal/http/middleware"
	"sms/backend/internal/http/response"
)

type Handler struct {
	svc    *modelsvc.Service
	logger *slog.Logger
}

func NewHandler(svc *modelsvc.Service, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.List)
	r.Get("/{id}", h.Detail)
	r.With(httpmiddleware.RequireRole("admin", "buyer")).Post("/", h.Create)
	r.With(httpmiddleware.RequireRole("admin", "buyer")).Put("/{id}", h.Update)
	r.With(httpmiddleware.RequireRole("admin", "buyer")).Delete("/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	filter := modelsvc.ListFilter{
		Page:       helpers.QueryInt(r, "page", 1),
		PageSize:   helpers.QueryInt(r, "pageSize", 10),
		Keyword:    r.URL.Query().Get("keyword"),
		SupplierID: func() *uuid.UUID { id := helpers.QueryUUID(r, "supplierId"); if id == uuid.Nil { return nil }; return &id }(),
		ModelType:  r.URL.Query().Get("modelType"),
		Status:     r.URL.Query().Get("status"),
	}
	result, err := h.svc.List(r.Context(), filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "查询失败")
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) Detail(w http.ResponseWriter, r *http.Request) {
	id := helpers.ParamUUID(r, "id")
	if id == uuid.Nil {
		response.Error(w, http.StatusBadRequest, "无效的 ID")
		return
	}
	m, err := h.svc.Get(r.Context(), id)
	if err != nil {
		helpers.HandleDomainError(w, err, h.logger)
		return
	}
	response.JSON(w, http.StatusOK, m)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input modelsvc.CreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	result, err := h.svc.Create(r.Context(), input)
	if err != nil {
		helpers.HandleDomainError(w, err, h.logger)
		return
	}
	response.JSON(w, http.StatusCreated, result)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := helpers.ParamUUID(r, "id")
	if id == uuid.Nil {
		response.Error(w, http.StatusBadRequest, "无效的 ID")
		return
	}
	var input modelsvc.UpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	result, err := h.svc.Update(r.Context(), id, input)
	if err != nil {
		helpers.HandleDomainError(w, err, h.logger)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := helpers.ParamUUID(r, "id")
	if id == uuid.Nil {
		response.Error(w, http.StatusBadRequest, "无效的 ID")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		helpers.HandleDomainError(w, err, h.logger)
		return
	}
	response.Void(w)
}
