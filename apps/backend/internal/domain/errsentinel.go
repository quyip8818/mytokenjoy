package domain

func NotFound(msg string) error {
	return NewDomainError(StatusNotFound, msg)
}

func Validation(msg string) error {
	return NewDomainError(StatusUnprocessable, msg)
}

func Forbidden(msg string) error {
	return NewDomainError(StatusForbidden, msg)
}

func BadRequest(msg string) error {
	return NewDomainError(StatusBadRequest, msg)
}
