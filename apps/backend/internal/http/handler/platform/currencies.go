package platform

import (
	"net/http"
	"regexp"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/http/response"
	"github.com/tokenjoy/backend/internal/http/httpx"
	"github.com/tokenjoy/backend/internal/store"
)

var currencyCodeRe = regexp.MustCompile(`^[A-Z]{3}$`)

// --- Catalog sync endpoint (public, no auth) ---

type catalogCurrencyDTO struct {
	Code         string `json:"code"`
	QuotaPerUnit int64  `json:"quotaPerUnit"`
}

// CatalogCurrencies returns enabled currencies for sync clients.
// GET /api/platform/sync/catalog/currencies (public)
func (h *Handler) CatalogCurrencies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	version, _ := h.p.SystemSettings.Get(ctx, catalogCurrenciesVersionKey)
	v, _ := strconv.Atoi(version)

	currencies, err := h.p.Billing.ListEnabledCurrencies(ctx)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	data := make([]catalogCurrencyDTO, 0, len(currencies))
	for _, c := range currencies {
		data = append(data, catalogCurrencyDTO{
			Code:         c.Code,
			QuotaPerUnit: c.QuotaPerUnit,
		})
	}
	response.JSON(w, http.StatusOK, map[string]any{"version": v, "data": data})
}

// --- Platform admin CRUD ---

type currencyResponse struct {
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

// ListCurrencies returns all currencies (enabled + disabled).
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

// CreateCurrency creates a new currency.
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

	// Check if exists already
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
	if err := h.p.Billing.UpsertCurrency(ctx, c); err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	h.p.SystemSettings.Increment(ctx, catalogCurrenciesVersionKey)

	// Re-read to get updatedAt
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

// UpdateCurrency updates quota_per_unit for an existing currency.
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

	existing.QuotaPerUnit = body.QuotaPerUnit
	existing.UpdatedByUserID = actorUserID(r)
	if err := h.p.Billing.UpsertCurrency(ctx, *existing); err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	h.p.SystemSettings.Increment(ctx, catalogCurrenciesVersionKey)

	updated, _ := h.p.Billing.GetCurrency(ctx, code)
	if updated == nil {
		response.JSON(w, http.StatusOK, toCurrencyResponse(*existing))
		return
	}
	response.JSON(w, http.StatusOK, toCurrencyResponse(*updated))
}

type toggleCurrencyStatusBody struct {
	Enabled bool `json:"enabled"`
}

// ToggleCurrencyStatus enables/disables a currency.
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

	h.p.SystemSettings.Increment(ctx, catalogCurrenciesVersionKey)

	updated, _ := h.p.Billing.GetCurrency(ctx, code)
	if updated == nil {
		existing.Enabled = body.Enabled
		response.JSON(w, http.StatusOK, toCurrencyResponse(*existing))
		return
	}
	response.JSON(w, http.StatusOK, toCurrencyResponse(*updated))
}
