package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"sms/backend/internal/config"
	authsvc "sms/backend/internal/domain/auth"
	httpmiddleware "sms/backend/internal/http/middleware"
	"sms/backend/internal/http/response"
)

type Handler struct {
	svc       *authsvc.Service
	cfg       config.Config
	jwtSecret string
}

func NewHandler(svc *authsvc.Service, cfg config.Config) *Handler {
	return &Handler{svc: svc, cfg: cfg, jwtSecret: cfg.JWTSecret}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/login", h.Login)
	r.Post("/refresh", h.Refresh)
	r.Post("/logout", h.Logout)
	r.With(httpmiddleware.Auth(h.jwtSecret)).Get("/profile", h.Profile)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	result, err := h.svc.Login(r.Context(), input.Username, input.Password)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "用户名或密码错误")
		return
	}
	h.setRefreshCookie(w, result.RefreshToken)
	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refreshToken")
	if err != nil || cookie.Value == "" {
		response.Error(w, http.StatusUnauthorized, "未登录或登录已过期")
		return
	}
	result, err := h.svc.Refresh(r.Context(), cookie.Value)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "登录已过期，请重新登录")
		return
	}
	h.setRefreshCookie(w, result.RefreshToken)
	response.JSON(w, http.StatusOK, map[string]string{"accessToken": result.AccessToken})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refreshToken")
	if err == nil && cookie.Value != "" {
		_ = h.svc.Logout(r.Context(), cookie.Value)
	}
	h.clearRefreshCookie(w)
	response.Void(w)
}

func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	user := httpmiddleware.UserFromCtx(r.Context())
	if user == nil {
		response.Error(w, http.StatusUnauthorized, "未登录")
		return
	}
	profile, err := h.svc.Profile(r.Context(), user.ID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "查询失败")
		return
	}
	response.JSON(w, http.StatusOK, profile)
}

func (h *Handler) setRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    token,
		Path:     "/api/auth",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.cfg.SecureCookie,
		MaxAge:   int(h.svc.RefreshTokenTTL() / time.Second),
	})
}

func (h *Handler) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    "",
		Path:     "/api/auth",
		HttpOnly: true,
		MaxAge:   -1,
	})
}
