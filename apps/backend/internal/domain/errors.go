package domain

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
	Message    string
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
