package httputil

import (
	"errors"
	"net/http"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/http/response"
)

const (
	MsgUnauthorized = "Unauthorized"
	MsgForbidden    = "Forbidden"
	MsgInternal     = "Internal server error"
	MsgBadBody      = "Invalid request body"
)

func WriteStatus(w http.ResponseWriter, status int, message string) {
	response.Error(w, status, message)
}

func WriteError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	var domainErr *domain.DomainError
	if errors.As(err, &domainErr) {
		if domainErr.RetryAfter != nil {
			response.JSON(w, domainErr.Status, map[string]any{
				"message":    domainErr.Message,
				"retryAfter": *domainErr.RetryAfter,
			})
			return
		}
		response.Error(w, domainErr.Status, domainErr.Message)
		return
	}
	response.Error(w, http.StatusInternalServerError, MsgInternal)
}
