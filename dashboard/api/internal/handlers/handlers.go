// Package handlers implements the HTTP handlers for the Fleet Inventory API.
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/COG-GTM/openSeaChest/dashboard/api/internal/models"
	"github.com/COG-GTM/openSeaChest/dashboard/api/internal/store"
)

// Handler holds dependencies for the API handlers.
type Handler struct {
	Store *store.Store
}

// New creates a Handler with the given store.
func New(s *store.Store) *Handler {
	return &Handler{Store: s}
}

// ReceiveReport handles POST /api/v1/agents/{agent_id}/report
func (h *Handler) ReceiveReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed", "")
		return
	}

	// Extract agent_id from path: /api/v1/agents/{agent_id}/report
	agentID := extractPathParam(r.URL.Path, "agents")
	if agentID == "" {
		writeError(w, http.StatusBadRequest, "missing agent_id in path", "")
		return
	}

	var req models.InventoryReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	// Override agent_id from path
	req.AgentID = agentID

	if req.Hostname == "" {
		writeError(w, http.StatusBadRequest, "hostname is required", "")
		return
	}

	// Upsert host
	hostID, err := h.Store.UpsertHost(req.AgentID, req.Hostname)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to register host", err.Error())
		return
	}

	// Upsert each device and create a snapshot
	for _, d := range req.Devices {
		if d.SerialNumber == "" {
			continue // skip devices without a serial
		}
		if err := h.Store.UpsertDevice(hostID, d); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to store device", err.Error())
			return
		}
		raw, _ := json.Marshal(d)
		if err := h.Store.InsertSnapshot(d.SerialNumber, string(raw)); err != nil {
			// snapshot failure is non-fatal; log and continue
			_ = err
		}
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "accepted",
		"host_id":      hostID,
		"device_count": len(req.Devices),
	})
}

// ListDevices handles GET /api/v1/devices
func (h *Handler) ListDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed", "")
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	perPage, _ := strconv.Atoi(q.Get("per_page"))

	filter := store.DeviceFilter{
		Host:        q.Get("host"),
		Model:       q.Get("model"),
		Interface:   q.Get("interface"),
		FirmwareRev: q.Get("firmware_rev"),
		Page:        page,
		PerPage:     perPage,
	}

	result, err := h.Store.ListDevices(filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list devices", err.Error())
		return
	}

	json.NewEncoder(w).Encode(result)
}

// GetDevice handles GET /api/v1/devices/{serial}
func (h *Handler) GetDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed", "")
		return
	}

	serial := extractPathParam(r.URL.Path, "devices")
	if serial == "" {
		writeError(w, http.StatusBadRequest, "missing serial number in path", "")
		return
	}

	device, err := h.Store.GetDeviceBySerial(serial)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get device", err.Error())
		return
	}
	if device == nil {
		writeError(w, http.StatusNotFound, "device not found", "")
		return
	}

	json.NewEncoder(w).Encode(device)
}

// ListHosts handles GET /api/v1/hosts
func (h *Handler) ListHosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed", "")
		return
	}

	hosts, err := h.Store.ListHosts()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list hosts", err.Error())
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"hosts": hosts,
	})
}

// ---------- helpers ----------

// extractPathParam extracts the path segment immediately after `key`.
// For /api/v1/agents/my-agent/report with key="agents", returns "my-agent".
func extractPathParam(path, key string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i, p := range parts {
		if p == key && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func writeError(w http.ResponseWriter, status int, msg, details string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.ErrorResponse{
		Error:   msg,
		Details: details,
	})
}
