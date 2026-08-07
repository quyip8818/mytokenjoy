package platform

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	domainredemption "github.com/tokenjoy/backend/internal/domain/redemption"
	"github.com/tokenjoy/backend/internal/http/httputil"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/store"
)

type generateCodesBody struct {
	BatchName     string  `json:"batchName"`
	FaceValue     float64 `json:"faceValue"`
	Quantity      int     `json:"quantity"`
	ExpiresInDays int     `json:"expiresInDays"`
	Note          string  `json:"note"`
}

func (h *Handler) GenerateRedemptionCodes(w http.ResponseWriter, r *http.Request) {
	var body generateCodesBody
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}

	operatorID := uuid.Nil
	if sessionCtx, ok := httpmiddleware.SessionFromContext(r.Context()); ok {
		operatorID = sessionCtx.Member.ID
	}

	result, err := h.p.RedemptionSvc.Generate(r.Context(), domainredemption.GenerateInput{
		BatchName:     body.BatchName,
		FaceValue:     body.FaceValue,
		Currency:      "", // default CNY
		Quantity:      body.Quantity,
		ExpiresInDays: body.ExpiresInDays,
		Note:          body.Note,
		CreatedBy:     operatorID,
	})
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, result, nil)
}

func (h *Handler) ListRedemptionCodes(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := store.RedemptionListFilter{
		Page:     parseInt(q.Get("page"), 1),
		PageSize: parseInt(q.Get("pageSize"), 20),
	}
	if bn := q.Get("batchName"); bn != "" {
		filter.BatchName = &bn
	}
	if s := q.Get("status"); s != "" {
		filter.Status = &s
	}

	result, err := h.p.RedemptionSvc.List(r.Context(), filter)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result, nil)
}

func parseInt(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 1 {
		return fallback
	}
	return v
}
