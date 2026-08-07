package platform

import (
	"net/http"
	"regexp"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/http/httpx"
	"github.com/tokenjoy/backend/internal/http/response"
	"github.com/tokenjoy/backend/internal/store"
)

var currencyCodeRe = regexp.MustCompile(`^[A-Z]{3}$`)

// --- Catalog sync endpoint (public, no auth) ---

type catalogCurrencyDTO struct {
	ID           string `json:"id"`
	Code         string `json:"code"`
	QuotaPerUnit int64  `json:"quotaPerUnit"`
	Enabled      bool   `json:"enabled"`
	UpdatedAt    int64  `json:"updatedAt"`
}

// CatalogCurrencies returns all currency rows (history) for sync clients.
// GET /api/platform/sync/catalog/currencies (public)
func (h *Handler) CatalogCurrencies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	v, _ := h.p.SyncVersions.Get(ctx, store.GlobalSyncVersion, "currencies")

	currencies, err := h.p.Billing.ListAllCurrencies(ctx)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	data := make([]catalogCurrencyDTO, 0, len(currencies))
	for _, c := range currencies {
		data = append(data, catalogCurrencyDTO{
			ID:           c.ID.String(),
			Code:         c.Code,
			QuotaPerUnit: c.QuotaPerUnit,
			Enabled:      c.Enabled,
			UpdatedAt:    c.UpdatedAt.Unix(),
		})
	}
	response.JSON(w, http.StatusOK, map[string]any{"version": v, "data": data})
}

// --- Platform admin CRUD ---

type currencyResponse struct {
	ID            string  `json:"id"`
	Code          string  `json:"code"`
	QuotaPerUnit  int64   `json:"quotaPerUnit"`
	Enabled       bool    `json:"enabled"`
	UpdatedAt     string  `json:"updatedAt"`
	UpdatedByName *string `json:"updatedByName"`
}

func toCurrencyResponse(c store.Currency) currencyResponse {
	var name *string
	if c.UpdatedByName != "" {
		name = &c.UpdatedByName
	}
	return currencyResponse{
		ID:            c.ID.String(),
		Code:          c.Code,
		QuotaPerUnit:  c.QuotaPerUnit,
		Enabled:       c.Enabled,
		UpdatedAt:     c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedByName: name,
	}
}

// actorUserID extracts the current user's ID from session context.
func actorUserID(r *http.Request) *uuid.UUID {
	session, ok := httpx.SessionFromContext(r.Context())
	if !ok {
		return nil
	}
	id := session.Member.UserID
	return &id
}

// ListCurrencies returns the latest row per currency (enabled + disabled).
// GET /api/platform/currencies
func (h *Handler) ListCurrencies(w http.ResponseWriter, r *http.Request) {
	currencies, err := h.p.Billing.ListAllCurrencies(r.Context())
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	out := make([]currencyResponse, 0, len(currencies))
	for _, c := range currencies {
		out = append(out, toCurrencyResponse(c))
	}
	response.JSON(w, http.StatusOK, out)
}

type createCurrencyBody struct {
	Code         string `json:"code"`
	QuotaPerUnit int64  `json:"quotaPerUnit"`
}

// CreateCurrency inserts a new currency row.
// POST /api/platform/currencies
func (h *Handler) CreateCurrency(w http.ResponseWriter, r *http.Request) {
	var body createCurrencyBody
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	if !currencyCodeRe.MatchString(body.Code) {
		httputil.WriteError(w, domain.BadRequest("currency code must be 3 uppercase letters"))
		return
	}
	if body.QuotaPerUnit <= 0 {
		httputil.WriteError(w, domain.BadRequest("quotaPerUnit must be positive"))
		return
	}

	ctx := r.Context()

	// Check if currency code already has any row
	existing, err := h.p.Billing.GetCurrency(ctx, body.Code)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	if existing != nil {
		httputil.WriteError(w, domain.Conflict("currency already exists"))
		return
	}

	c := store.Currency{Code: body.Code, QuotaPerUnit: body.QuotaPerUnit, Enabled: true, UpdatedByUserID: actorUserID(r)}
	if err := h.p.Billing.InsertCurrency(ctx, c); err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	h.p.SyncVersions.Increment(ctx, store.GlobalSyncVersion, "currencies")

	// Re-read to get id + updatedAt
	created, _ := h.p.Billing.GetCurrency(ctx, body.Code)
	if created == nil {
		response.JSON(w, http.StatusCreated, toCurrencyResponse(c))
		return
	}
	response.JSON(w, http.StatusCreated, toCurrencyResponse(*created))
}

type updateCurrencyBody struct {
	QuotaPerUnit int64 `json:"quotaPerUnit"`
}

// UpdateCurrency inserts a new row with updated quota_per_unit.
// PUT /api/platform/currencies/{code}
func (h *Handler) UpdateCurrency(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if !currencyCodeRe.MatchString(code) {
		httputil.WriteError(w, domain.BadRequest("invalid currency code"))
		return
	}

	var body updateCurrencyBody
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	if body.QuotaPerUnit <= 0 {
		httputil.WriteError(w, domain.BadRequest("quotaPerUnit must be positive"))
		return
	}

	ctx := r.Context()
	existing, err := h.p.Billing.GetCurrency(ctx, code)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	if existing == nil {
		httputil.WriteError(w, domain.NotFound("currency not found"))
		return
	}

	// Insert new row with updated QPU
	c := store.Currency{
		Code:            code,
		QuotaPerUnit:    body.QuotaPerUnit,
		Enabled:         existing.Enabled,
		UpdatedByUserID: actorUserID(r),
	}
	if err := h.p.Billing.InsertCurrency(ctx, c); err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	h.p.SyncVersions.Increment(ctx, store.GlobalSyncVersion, "currencies")

	updated, _ := h.p.Billing.GetCurrency(ctx, code)
	if updated == nil {
		response.JSON(w, http.StatusOK, toCurrencyResponse(c))
		return
	}
	response.JSON(w, http.StatusOK, toCurrencyResponse(*updated))
}

type toggleCurrencyStatusBody struct {
	Enabled bool `json:"enabled"`
}

// ToggleCurrencyStatus inserts a new row with updated enabled status.
// PATCH /api/platform/currencies/{code}/status
func (h *Handler) ToggleCurrencyStatus(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if !currencyCodeRe.MatchString(code) {
		httputil.WriteError(w, domain.BadRequest("invalid currency code"))
		return
	}

	var body toggleCurrencyStatusBody
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}

	ctx := r.Context()
	existing, err := h.p.Billing.GetCurrency(ctx, code)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	if existing == nil {
		httputil.WriteError(w, domain.NotFound("currency not found"))
		return
	}

	// If disabling, check FK references
	if !body.Enabled {
		referenced, err := h.p.Billing.IsCurrencyReferenced(ctx, code)
		if err != nil {
			httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
			return
		}
		if referenced {
			httputil.WriteError(w, domain.Conflict("currency is referenced by companies"))
			return
		}
	}

	if err := h.p.Billing.SetCurrencyEnabled(ctx, code, body.Enabled, actorUserID(r)); err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	h.p.SyncVersions.Increment(ctx, store.GlobalSyncVersion, "currencies")

	updated, _ := h.p.Billing.GetCurrency(ctx, code)
	if updated == nil {
		existing.Enabled = body.Enabled
		response.JSON(w, http.StatusOK, toCurrencyResponse(*existing))
		return
	}
	response.JSON(w, http.StatusOK, toCurrencyResponse(*updated))
}

// ListCurrencyHistory returns all historical rows for a currency code.
// GET /api/platform/currencies/{code}/history?limit=50&offset=0
func (h *Handler) ListCurrencyHistory(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if !currencyCodeRe.MatchString(code) {
		httputil.WriteError(w, domain.BadRequest("invalid currency code"))
		return
	}

	limit := 50
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	rows, err := h.p.Billing.ListCurrencyHistory(r.Context(), code, limit, offset)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	out := make([]currencyResponse, 0, len(rows))
	for _, c := range rows {
		out = append(out, toCurrencyResponse(c))
	}
	response.JSON(w, http.StatusOK, out)
}
