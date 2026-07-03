package shared

import (
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
)

type ProtectedHandlerBase struct {
	httpdeps.Protected
}

func NewProtectedHandlerBase(p httpdeps.Protected) ProtectedHandlerBase {
	return ProtectedHandlerBase{Protected: p}
}
