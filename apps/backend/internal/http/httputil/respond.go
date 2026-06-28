package httputil

import (
	"net/http"

	"github.com/tokenjoy/backend/internal/http/response"
)

func WriteJSON(w http.ResponseWriter, status int, v any, err error) {
	if err != nil {
		WriteError(w, err)
		return
	}
	response.JSON(w, status, v)
}

func WriteOK(w http.ResponseWriter, v any) {
	response.JSON(w, http.StatusOK, v)
}

func WriteVoid(w http.ResponseWriter, err error) {
	if err != nil {
		WriteError(w, err)
		return
	}
	response.Void(w)
}
