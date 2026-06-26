package domain

const (
	StatusBadRequest    = 400
	StatusNotFound      = 404
	StatusUnprocessable = 422
)

type DomainError struct {
	Status  int
	Message string
}

func (e *DomainError) Error() string {
	return e.Message
}

func NewDomainError(status int, message string) *DomainError {
	return &DomainError{Status: status, Message: message}
}
