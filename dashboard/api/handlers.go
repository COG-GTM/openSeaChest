package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/COG-GTM/openSeaChest/dashboard/models"
)

// Handler holds the dependencies for the REST API handlers.
type Handler struct {
	store *Store
}

// NewHandler creates a new Handler with the given store.
func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

// RegisterRoutes registers all API routes on the given router.
func (h *Handler) RegisterRoutes(r *mux.Router) {
	api := r.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/devices/{serial}/health", h.GetDeviceHealth).Methods("GET")
	api.HandleFunc("/devices/{serial}/health/history", h.GetDeviceHealthHistory).Methods("GET")
	api.HandleFunc("/fleet/health/summary", h.GetFleetHealthSummary).Methods("GET")
}

// GetDeviceHealth handles GET /api/v1/devices/{serial}/health
// Returns the current health summary for a specific device.
func (h *Handler) GetDeviceHealth(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serial := vars["serial"]

	if serial == "" {
		writeError(w, http.StatusBadRequest, "device serial is required")
		return
	}

	summary, err := h.store.GetDeviceHealthSummary(r.Context(), serial)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to retrieve device health: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

// GetDeviceHealthHistory handles GET /api/v1/devices/{serial}/health/history
// Query parameters:
//   - from: RFC3339 start time (required)
//   - to:   RFC3339 end time (required)
//   - metric: metric name filter (optional)
func (h *Handler) GetDeviceHealthHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serial := vars["serial"]

	if serial == "" {
		writeError(w, http.StatusBadRequest, "device serial is required")
		return
	}

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	metricName := r.URL.Query().Get("metric")

	if fromStr == "" || toStr == "" {
		writeError(w, http.StatusBadRequest, "'from' and 'to' query parameters are required (RFC3339 format)")
		return
	}

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid 'from' parameter: must be RFC3339 format")
		return
	}

	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid 'to' parameter: must be RFC3339 format")
		return
	}

	if to.Before(from) {
		writeError(w, http.StatusBadRequest, "'to' must be after 'from'")
		return
	}

	dataPoints, err := h.store.GetHealthHistory(r.Context(), serial, from, to, metricName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to retrieve health history: "+err.Error())
		return
	}

	response := models.HealthHistoryResponse{
		DeviceSerial: serial,
		From:         from,
		To:           to,
		MetricName:   metricName,
		DataPoints:   dataPoints,
	}

	writeJSON(w, http.StatusOK, response)
}

// GetFleetHealthSummary handles GET /api/v1/fleet/health/summary
// Returns fleet-wide rollup showing healthy/warning/critical/unknown counts.
func (h *Handler) GetFleetHealthSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.store.GetFleetHealthSummary(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to retrieve fleet health summary: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

// writeJSON marshals the given value as JSON and writes it to the response.
func writeJSON(w http.ResponseWriter, statusCode int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// writeError writes a standard API error response.
func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, models.APIError{
		Code:    statusCode,
		Message: message,
	})
}
