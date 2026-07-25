package types

import "errors"

var (
	ErrNotFound     = errors.New("record not found")
	ErrConflict     = errors.New("record already exists")
	ErrHasRefs      = errors.New("record has references")
	ErrValidation   = errors.New("validation error")
	ErrUnauthorized = errors.New("unauthorized")
)
