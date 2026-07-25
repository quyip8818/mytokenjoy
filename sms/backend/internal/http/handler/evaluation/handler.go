package evaluation

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	evalsvc "sms/backend/internal/domain/evaluation"
	"sms/backend/internal/domain/types"
	"sms/backend/internal/http/helpers"
	httpmiddleware "sms/backend/internal/http/middleware"
	"sms/backend/internal/http/response"
)

type Handler struct {
	svc    *evalsvc.Service
	logger *slog.Logger
}

func NewHandler(svc *evalsvc.Service, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.List)
	r.Get("/weights", h.GetWeights)
	r.Get("/{id}", h.Detail)
	r.With(httpmiddleware.RequireRole("admin", "buyer")).Post("/", h.Create)
	r.With(httpmiddleware.RequireRole("admin", "buyer")).Put("/{id}", h.Update)
	r.With(httpmiddleware.RequireRole("admin", "buyer")).Delete("/{id}", h.Delete)
	r.With(httpmiddleware.RequireRole("admin")).Put("/weights", h.UpdateWeights)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	filter := evalsvc.ListFilter{
		Page:       helpers.QueryInt(r, "page", 1),
		PageSize:   helpers.QueryInt(r, "pageSize", 10),
		SupplierID: helpers.QueryUUID(r, "supplierId"),
		Period:     r.URL.Query().Get("period"),
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
	e, err := h.svc.Get(r.Context(), id)
	if err != nil {
		helpers.HandleDomainError(w, err, h.logger)
		return
	}
	response.JSON(w, http.StatusOK, e)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input evalsvc.CreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	user := httpmiddleware.UserFromCtx(r.Context())
	result, err := h.svc.Create(r.Context(), user.ID, input)
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
	var input evalsvc.UpdateInput
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

func (h *Handler) GetWeights(w http.ResponseWriter, r *http.Request) {
	weights, err := h.svc.GetWeights(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "查询失败")
		return
	}
	response.JSON(w, http.StatusOK, weights)
}

func (h *Handler) UpdateWeights(w http.ResponseWriter, r *http.Request) {
	var weights []types.EvaluationWeight
	if err := json.NewDecoder(r.Body).Decode(&weights); err != nil {
		response.Error(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	if err := h.svc.UpdateWeights(r.Context(), weights); err != nil {
		helpers.HandleDomainError(w, err, h.logger)
		return
	}
	response.Void(w)
}
