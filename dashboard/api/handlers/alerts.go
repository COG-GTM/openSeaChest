// Package handlers implements HTTP request handlers for the alerting API.
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/COG-GTM/openSeaChest/dashboard/api/models"
	"github.com/COG-GTM/openSeaChest/dashboard/api/services"
)

// AlertHandler handles HTTP requests for alert operations.
type AlertHandler struct {
	store *services.Store
}

// NewAlertHandler creates a new AlertHandler.
func NewAlertHandler(store *services.Store) *AlertHandler {
	return &AlertHandler{store: store}
}

// HandleActiveAlerts handles GET /api/v1/alerts/active.
func (h *AlertHandler) HandleActiveAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	alerts := h.store.ListActiveAlerts()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// HandleAlertHistory handles GET /api/v1/alerts/history.
func (h *AlertHandler) HandleAlertHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	page := 1
	pageSize := 50

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	alerts, total := h.store.ListAlertHistory(page, pageSize)
	resp := models.AlertListResponse{
		Alerts:     alerts,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}
	writeJSON(w, http.StatusOK, resp)
}

// HandleAcknowledge handles POST /api/v1/alerts/{id}/acknowledge.
func (h *AlertHandler) HandleAcknowledge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Extract alert ID: path is /api/v1/alerts/{id}/acknowledge
	id := extractAlertIDFromAction(r.URL.Path, "acknowledge")
	if id == "" {
		writeError(w, http.StatusBadRequest, "alert ID is required")
		return
	}

	var req models.AcknowledgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Allow empty body, default to "system"
		req.AcknowledgedBy = "system"
	}
	if req.AcknowledgedBy == "" {
		req.AcknowledgedBy = "system"
	}

	alert, err := h.store.AcknowledgeAlert(id, req.AcknowledgedBy)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusConflict, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, alert)
}

// HandleResolve handles POST /api/v1/alerts/{id}/resolve.
func (h *AlertHandler) HandleResolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Extract alert ID: path is /api/v1/alerts/{id}/resolve
	id := extractAlertIDFromAction(r.URL.Path, "resolve")
	if id == "" {
		writeError(w, http.StatusBadRequest, "alert ID is required")
		return
	}

	var req models.ResolveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.ResolvedBy = "system"
	}
	if req.ResolvedBy == "" {
		req.ResolvedBy = "system"
	}

	alert, err := h.store.ResolveAlert(id, req.ResolvedBy)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusConflict, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, alert)
}

// HandleEvaluate handles POST /api/v1/alerts/evaluate for testing the rule engine.
func (h *AlertHandler) HandleEvaluate(w http.ResponseWriter, r *http.Request, engine *services.RuleEngine) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var report services.DeviceReport
	if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	engine.Evaluate(report)
	writeJSON(w, http.StatusOK, map[string]string{"status": "evaluated"})
}

// extractAlertIDFromAction extracts the alert ID from paths like
// /api/v1/alerts/{id}/acknowledge or /api/v1/alerts/{id}/resolve.
func extractAlertIDFromAction(path, action string) string {
	// Remove trailing slash
	path = strings.TrimSuffix(path, "/")
	// Expected: /api/v1/alerts/{id}/{action}
	suffix := "/" + action
	if !strings.HasSuffix(path, suffix) {
		return ""
	}
	path = strings.TrimSuffix(path, suffix)
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}
