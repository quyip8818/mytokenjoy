package response

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v == nil {
		_, _ = w.Write([]byte("null"))
		return
	}
	_ = json.NewEncoder(w).Encode(v)
}

func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{"message": message})
}

func Void(w http.ResponseWriter) {
	JSON(w, http.StatusOK, nil)
}
