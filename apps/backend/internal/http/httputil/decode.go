package httputil

import (
	"encoding/json"
	"net/http"

	"github.com/tokenjoy/backend/internal/domain"
)

// maxRequestBodySize limits JSON request bodies to 1 MB.
const maxRequestBodySize = 1 << 20

func DecodeJSON(r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, maxRequestBodySize)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return domain.BadRequest(MsgBadBody)
	}
	return nil
}
