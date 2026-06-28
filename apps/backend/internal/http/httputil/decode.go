package httputil

import (
	"encoding/json"
	"net/http"

	"github.com/tokenjoy/backend/internal/domain"
)

func DecodeJSON(r *http.Request, dst any) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return domain.BadRequest(MsgBadBody)
	}
	return nil
}
