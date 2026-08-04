package domain

import "errors"

const (
	StatusBadRequest         = 400
	StatusNotFound           = 404
	StatusForbidden          = 403
	StatusConflict           = 409
	StatusUnprocessable      = 422
	StatusTooManyRequests    = 429
	StatusNotImplemented     = 501
	StatusServiceUnavailable = 503
)

type DomainError struct {
	Status     int
	Code       string         // 机器可读错误码，空串 = 不输出
	Message    string
	Meta       map[string]any // 可选结构化上下文，nil = 不输出
	RetryAfter *int
}

func (e *DomainError) Error() string {
	return e.Message
}

func NewDomainError(status int, message string) *DomainError {
	return &DomainError{Status: status, Message: message}
}

func NewDomainErrorWithRetryAfter(status int, message string, retryAfter int) *DomainError {
	return &DomainError{Status: status, Message: message, RetryAfter: &retryAfter}
}

func IsServiceUnavailable(err error) bool {
	var domainErr *DomainError
	if !errors.As(err, &domainErr) {
		return false
	}
	return domainErr.Status == StatusServiceUnavailable
}
