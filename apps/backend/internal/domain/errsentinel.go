package domain

import "errors"

func NotFound(msg string) error {
	return NewDomainError(StatusNotFound, msg)
}

func Validation(msg string) error {
	return NewDomainError(StatusUnprocessable, msg)
}

// ValidationCode creates a 422 error with a machine-readable code and optional meta.
func ValidationCode(code, msg string, meta ...map[string]any) error {
	e := &DomainError{Status: StatusUnprocessable, Code: code, Message: msg}
	if len(meta) > 0 {
		e.Meta = meta[0]
	}
	return e
}

func Forbidden(msg string) error {
	return NewDomainError(StatusForbidden, msg)
}

// ForbiddenCode creates a 403 error with a machine-readable code and optional meta.
func ForbiddenCode(code, msg string, meta ...map[string]any) error {
	e := &DomainError{Status: StatusForbidden, Code: code, Message: msg}
	if len(meta) > 0 {
		e.Meta = meta[0]
	}
	return e
}

func Conflict(msg string) error {
	return NewDomainError(StatusConflict, msg)
}

func BadRequest(msg string) error {
	return NewDomainError(StatusBadRequest, msg)
}

func ServiceUnavailable(msg string) error {
	return NewDomainError(StatusServiceUnavailable, msg)
}

func TooManyRequests(msg string) error {
	return NewDomainError(StatusTooManyRequests, msg)
}

func IsNotFound(err error) bool {
	var domainErr *DomainError
	if !errors.As(err, &domainErr) {
		return false
	}
	return domainErr.Status == StatusNotFound
}
