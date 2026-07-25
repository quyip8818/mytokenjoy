package helpers

import (
	"errors"
	"log/slog"
	"net/http"

	"sms/backend/internal/domain/types"
	"sms/backend/internal/http/response"
)

// HandleDomainError maps domain errors to HTTP responses.
// It covers the union of all error cases across all handlers.
func HandleDomainError(w http.ResponseWriter, err error, logger *slog.Logger) {
	switch {
	case errors.Is(err, types.ErrNotFound):
		response.Error(w, http.StatusNotFound, "记录不存在")
	case errors.Is(err, types.ErrConflict):
		response.Error(w, http.StatusConflict, "记录已存在")
	case errors.Is(err, types.ErrHasRefs):
		response.Error(w, http.StatusConflict, err.Error())
	case errors.Is(err, types.ErrValidation):
		response.Error(w, http.StatusBadRequest, err.Error())
	default:
		logger.Error("unhandled handler error", "err", err)
		response.Error(w, http.StatusInternalServerError, "内部错误")
	}
}
