package handler

import (
	"errors"
	"net/http"

	"github.com/tokenjoy/backend/internal/domain"
	"github.com/tokenjoy/backend/internal/http/response"
)

func writeDomainError(w http.ResponseWriter, err error) bool {
	var domainErr *domain.DomainError
	if errors.As(err, &domainErr) {
		if domainErr.RetryAfter != nil {
			response.JSON(w, domainErr.Status, map[string]any{
				"message":    domainErr.Message,
				"retryAfter": *domainErr.RetryAfter,
			})
			return true
		}
		response.Error(w, domainErr.Status, domainErr.Message)
		return true
	}
	return false
}
