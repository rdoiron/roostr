package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

// ErrorResponse represents a structured API error.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details any    `json:"details,omitempty"`
}

// respondJSON sends a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("Error encoding JSON response: %v", err)
		}
	}
}

// respondError sends a structured error response.
func respondError(w http.ResponseWriter, status int, message, code string) {
	respondJSON(w, status, ErrorResponse{
		Error: message,
		Code:  code,
	})
}

// respondErrorWithDetails sends a structured error response with additional details.
func respondErrorWithDetails(w http.ResponseWriter, status int, message, code string, details any) {
	respondJSON(w, status, ErrorResponse{
		Error:   message,
		Code:    code,
		Details: details,
	})
}
