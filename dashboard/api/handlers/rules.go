// Package handlers implements HTTP request handlers for the alerting API.
package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/COG-GTM/openSeaChest/dashboard/api/models"
	"github.com/COG-GTM/openSeaChest/dashboard/api/services"
)

// RuleHandler handles HTTP requests for alert rule CRUD operations.
type RuleHandler struct {
	store *services.Store
}

// NewRuleHandler creates a new RuleHandler.
func NewRuleHandler(store *services.Store) *RuleHandler {
	return &RuleHandler{store: store}
}

// HandleRules dispatches requests for /api/v1/alerts/rules.
func (h *RuleHandler) HandleRules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createRule(w, r)
	case http.MethodGet:
		h.listRules(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// HandleRuleByID dispatches requests for /api/v1/alerts/rules/{id}.
func (h *RuleHandler) HandleRuleByID(w http.ResponseWriter, r *http.Request) {
	id := extractIDFromPath(r.URL.Path, "/api/v1/alerts/rules/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "rule ID is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getRule(w, id)
	case http.MethodPut:
		h.updateRule(w, r, id)
	case http.MethodDelete:
		h.deleteRule(w, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *RuleHandler) createRule(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	rule, err := h.store.CreateRule(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, rule)
}

func (h *RuleHandler) listRules(w http.ResponseWriter, _ *http.Request) {
	rules := h.store.ListRules()
	resp := models.RuleListResponse{Rules: rules}
	writeJSON(w, http.StatusOK, resp)
}

func (h *RuleHandler) getRule(w http.ResponseWriter, id string) {
	rule, err := h.store.GetRule(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rule)
}

func (h *RuleHandler) updateRule(w http.ResponseWriter, r *http.Request, id string) {
	var req models.UpdateRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	rule, err := h.store.UpdateRule(id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
		} else {
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, rule)
}

func (h *RuleHandler) deleteRule(w http.ResponseWriter, id string) {
	if err := h.store.DeleteRule(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
