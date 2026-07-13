package domain

import "errors"

func NotFound(msg string) error {
	return NewDomainError(StatusNotFound, msg)
}

func Validation(msg string) error {
	return NewDomainError(StatusUnprocessable, msg)
}

func Forbidden(msg string) error {
	return NewDomainError(StatusForbidden, msg)
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
