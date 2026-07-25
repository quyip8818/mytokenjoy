package contract

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	contractsvc "sms/backend/internal/domain/contract"
	"sms/backend/internal/http/helpers"
	httpmiddleware "sms/backend/internal/http/middleware"
	"sms/backend/internal/http/response"
)

type Handler struct {
	svc    *contractsvc.Service
	logger *slog.Logger
}

func NewHandler(svc *contractsvc.Service, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.List)
	r.Get("/{id}", h.Detail)
	r.With(httpmiddleware.RequireRole("admin", "buyer")).Post("/", h.Create)
	r.With(httpmiddleware.RequireRole("admin", "buyer")).Put("/{id}", h.Update)
	r.With(httpmiddleware.RequireRole("admin", "buyer")).Delete("/{id}", h.Delete)
	r.With(httpmiddleware.RequireRole("admin", "buyer")).Post("/{id}/attachments", h.Upload)
	r.Get("/{id}/attachments/{aid}/download", h.Download)
	r.With(httpmiddleware.RequireRole("admin", "buyer")).Delete("/{id}/attachments/{aid}", h.DeleteAttachment)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	filter := contractsvc.ListFilter{
		Page:       helpers.QueryInt(r, "page", 1),
		PageSize:   helpers.QueryInt(r, "pageSize", 10),
		Keyword:    r.URL.Query().Get("keyword"),
		SupplierID: helpers.QueryUUID(r, "supplierId"),
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
	detail, err := h.svc.Get(r.Context(), id)
	if err != nil {
		helpers.HandleDomainError(w, err, h.logger)
		return
	}
	response.JSON(w, http.StatusOK, detail)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input contractsvc.CreateInput
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
	var input contractsvc.UpdateInput
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

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	id := helpers.ParamUUID(r, "id")
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		response.Error(w, http.StatusBadRequest, "文件过大或格式错误")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "未找到上传文件")
		return
	}
	defer file.Close()
	user := httpmiddleware.UserFromCtx(r.Context())
	att, err := h.svc.UploadAttachment(r.Context(), id, user.ID, header, file)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "上传失败")
		return
	}
	response.JSON(w, http.StatusCreated, att)
}

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	aid := helpers.ParamUUID(r, "aid")
	att, err := h.svc.GetAttachment(r.Context(), aid)
	if err != nil {
		helpers.HandleDomainError(w, err, h.logger)
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename=\""+att.FileName+"\"")
	http.ServeFile(w, r, att.FilePath)
}

func (h *Handler) DeleteAttachment(w http.ResponseWriter, r *http.Request) {
	aid := helpers.ParamUUID(r, "aid")
	if err := h.svc.DeleteAttachment(r.Context(), aid); err != nil {
		helpers.HandleDomainError(w, err, h.logger)
		return
	}
	response.Void(w)
}
