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
		body := map[string]any{"message": domainErr.Message}
		if domainErr.Code != "" {
			body["code"] = domainErr.Code
		}
		if domainErr.Meta != nil {
			body["meta"] = domainErr.Meta
		}
		if domainErr.RetryAfter != nil {
			body["retryAfter"] = *domainErr.RetryAfter
		}
		response.JSON(w, domainErr.Status, body)
		return
	}
	response.Error(w, http.StatusInternalServerError, MsgInternal)
}
