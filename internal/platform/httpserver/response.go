package httpserver

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, r *http.Request, statusCode int, body map[string]any) {
	if body == nil {
		body = map[string]any{}
	}

	if _, exists := body["correlation_id"]; !exists {
		body["correlation_id"] = CorrelationIDFromContext(r.Context())
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func WriteData(w http.ResponseWriter, r *http.Request, statusCode int, data any) {
	WriteJSON(w, r, statusCode, map[string]any{
		"data": data,
	})
}

func DecodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(dst)
}
