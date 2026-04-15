// Package handlers implements the HTTP request handlers for the firmware
// compliance auditing API.
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/COG-GTM/openSeaChest/dashboard/api/models"
)

// PolicyHandler holds the database connection used by policy endpoints.
type PolicyHandler struct {
	DB *sql.DB
}

// NewPolicyHandler creates a PolicyHandler with the given database connection.
func NewPolicyHandler(db *sql.DB) *PolicyHandler {
	return &PolicyHandler{DB: db}
}

// CreatePolicy handles POST /api/v1/firmware/policies.
// It creates a new firmware policy from the JSON request body.
func (h *PolicyHandler) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	var input models.FirmwarePolicyInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if input.Name == "" || input.ModelPattern == "" || input.RequiredFirmware == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "name, model_pattern, and required_firmware are required",
		})
		return
	}

	if _, err := filepath.Match(input.ModelPattern, ""); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "model_pattern is not a valid glob pattern"})
		return
	}

	severity := input.Severity
	if severity == "" {
		severity = "warning"
	}
	if severity != "critical" && severity != "warning" && severity != "info" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "severity must be critical, warning, or info"})
		return
	}

	now := time.Now().UTC()
	result, err := h.DB.Exec(
		`INSERT INTO firmware_policies (name, model_pattern, required_firmware, severity, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		input.Name, input.ModelPattern, input.RequiredFirmware, severity, now, now,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create policy"})
		return
	}

	id, _ := result.LastInsertId()
	policy := models.FirmwarePolicy{
		ID:               id,
		Name:             input.Name,
		ModelPattern:     input.ModelPattern,
		RequiredFirmware: input.RequiredFirmware,
		Severity:         severity,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	writeJSON(w, http.StatusCreated, policy)
}

// ListPolicies handles GET /api/v1/firmware/policies.
// It returns all firmware policies ordered by creation time.
func (h *PolicyHandler) ListPolicies(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(
		`SELECT id, name, model_pattern, required_firmware, severity, created_at, updated_at
		 FROM firmware_policies ORDER BY created_at DESC`,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list policies"})
		return
	}
	defer rows.Close()

	policies := make([]models.FirmwarePolicy, 0)
	for rows.Next() {
		var p models.FirmwarePolicy
		if err := rows.Scan(&p.ID, &p.Name, &p.ModelPattern, &p.RequiredFirmware, &p.Severity, &p.CreatedAt, &p.UpdatedAt); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to scan policy row"})
			return
		}
		policies = append(policies, p)
	}

	if err := rows.Err(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to iterate policy rows"})
		return
	}

	writeJSON(w, http.StatusOK, policies)
}

// UpdatePolicy handles PUT /api/v1/firmware/policies/{id}.
// It updates an existing firmware policy identified by the URL path parameter.
func (h *PolicyHandler) UpdatePolicy(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid policy id"})
		return
	}

	var input models.FirmwarePolicyInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Verify the policy exists.
	var existing models.FirmwarePolicy
	err = h.DB.QueryRow(
		`SELECT id, name, model_pattern, required_firmware, severity, created_at, updated_at
		 FROM firmware_policies WHERE id = ?`, id,
	).Scan(&existing.ID, &existing.Name, &existing.ModelPattern, &existing.RequiredFirmware,
		&existing.Severity, &existing.CreatedAt, &existing.UpdatedAt)
	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "policy not found"})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch policy"})
		return
	}

	// Apply partial updates: only overwrite non-empty fields.
	if input.Name != "" {
		existing.Name = input.Name
	}
	if input.ModelPattern != "" {
		if _, err := filepath.Match(input.ModelPattern, ""); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "model_pattern is not a valid glob pattern"})
			return
		}
		existing.ModelPattern = input.ModelPattern
	}
	if input.RequiredFirmware != "" {
		existing.RequiredFirmware = input.RequiredFirmware
	}
	if input.Severity != "" {
		if input.Severity != "critical" && input.Severity != "warning" && input.Severity != "info" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "severity must be critical, warning, or info"})
			return
		}
		existing.Severity = input.Severity
	}

	now := time.Now().UTC()
	existing.UpdatedAt = now

	_, err = h.DB.Exec(
		`UPDATE firmware_policies SET name=?, model_pattern=?, required_firmware=?, severity=?, updated_at=?
		 WHERE id=?`,
		existing.Name, existing.ModelPattern, existing.RequiredFirmware, existing.Severity, now, id,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update policy"})
		return
	}

	writeJSON(w, http.StatusOK, existing)
}

// writeJSON is a helper that serialises v as JSON and writes it to w with the
// given HTTP status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

// isTableNotFound returns true if the error indicates a missing SQLite table
// (e.g. "no such table: devices"). This is used to distinguish expected
// missing-table errors from real database failures.
func isTableNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "no such table")
}
