package response

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error,omitempty"`
}

func WriteError(w http.ResponseWriter, code int, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Error: err})
}

func WriteJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func WriteOK(w http.ResponseWriter, data any) {
	WriteJSON(w, http.StatusOK, data)
}

func WriteCreated(w http.ResponseWriter, data any) {
	WriteJSON(w, http.StatusCreated, data)
}

func WriteBadRequest(w http.ResponseWriter, err string) {
	WriteError(w, http.StatusBadRequest, err)
}

func WriteUnauthorized(w http.ResponseWriter, err string) {
	WriteError(w, http.StatusUnauthorized, err)
}

func WriteInternalError(w http.ResponseWriter, err string) {
	WriteError(w, http.StatusInternalServerError, err)
}

// Re-export commonly used types for convenience
type ResponseWriter = http.ResponseWriter
type Request = http.Request