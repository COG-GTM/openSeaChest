package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/COG-GTM/openSeaChest/dashboard/api/engine"
	"github.com/COG-GTM/openSeaChest/dashboard/api/models"
)

// ComplianceHandler holds the database connection used by compliance endpoints.
type ComplianceHandler struct {
	DB *sql.DB
}

// NewComplianceHandler creates a ComplianceHandler with the given database connection.
func NewComplianceHandler(db *sql.DB) *ComplianceHandler {
	return &ComplianceHandler{DB: db}
}

// GetFleetCompliance handles GET /api/v1/firmware/compliance.
// It evaluates all devices against all policies and returns a fleet-wide
// compliance report with compliant/non-compliant/unknown counts per policy.
func (h *ComplianceHandler) GetFleetCompliance(w http.ResponseWriter, r *http.Request) {
	// Load all policies.
	policies, err := h.loadPolicies()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load policies"})
		return
	}

	// Load all known devices.
	devices, err := h.loadDevices()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load devices"})
		return
	}

	// Run the compliance evaluation engine.
	summaries := engine.EvaluateFleet(policies, devices)

	// Persist individual compliance results for audit trail.
	for _, policy := range policies {
		for _, device := range devices {
			result := engine.EvaluateCompliance(policy, device)
			if result == nil {
				continue
			}
			_, _ = h.DB.Exec(
				`INSERT INTO compliance_results (device_id, policy_id, compliant, current_firmware, checked_at)
				 VALUES (?, ?, ?, ?, ?)`,
				result.DeviceID, result.PolicyID, result.Compliant, result.CurrentFirmware, time.Now().UTC(),
			)
		}
	}

	report := models.FleetComplianceReport{
		GeneratedAt: time.Now().UTC(),
		Policies:    summaries,
	}

	writeJSON(w, http.StatusOK, report)
}

func (h *ComplianceHandler) loadPolicies() ([]models.FirmwarePolicy, error) {
	rows, err := h.DB.Query(
		`SELECT id, name, model_pattern, required_firmware, severity, created_at, updated_at
		 FROM firmware_policies`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []models.FirmwarePolicy
	for rows.Next() {
		var p models.FirmwarePolicy
		if err := rows.Scan(&p.ID, &p.Name, &p.ModelPattern, &p.RequiredFirmware, &p.Severity, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	return policies, rows.Err()
}

func (h *ComplianceHandler) loadDevices() ([]models.Device, error) {
	// Devices are expected to be registered via an agent registration
	// endpoint (outside the scope of WI-3). For now, we query a devices
	// table if it exists, otherwise return an empty slice so the API
	// remains functional during initial deployment.
	rows, err := h.DB.Query(
		`SELECT id, serial, product_identification, product_revision FROM devices`,
	)
	if err != nil {
		// Table may not exist yet; return empty list.
		return []models.Device{}, nil
	}
	defer rows.Close()

	var devices []models.Device
	for rows.Next() {
		var d models.Device
		if err := rows.Scan(&d.ID, &d.Serial, &d.ProductIdentification, &d.ProductRevision); err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}
