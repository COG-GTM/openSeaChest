// Package handlers implements HTTP request handlers for the alerting API.
package handlers

import "net/http"

// HandleNotFound returns a 404 JSON response.
func HandleNotFound(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusNotFound, "endpoint not found")
}
