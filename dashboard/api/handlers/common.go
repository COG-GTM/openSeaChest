// Package handlers implements HTTP request handlers for the alerting API.
package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

// ErrorResponse is the standard error response format.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}

// extractIDFromPath extracts a resource ID from a URL path given a prefix.
// For example, extractIDFromPath("/api/v1/alerts/rules/abc-123", "/api/v1/alerts/rules/")
// returns "abc-123".
func extractIDFromPath(path, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	id := strings.TrimPrefix(path, prefix)
	id = strings.TrimSuffix(id, "/")
	// If the ID contains further path segments, only take the first one
	if idx := strings.Index(id, "/"); idx >= 0 {
		id = id[:idx]
	}
	return id
}
