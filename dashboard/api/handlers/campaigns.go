package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/COG-GTM/openSeaChest/dashboard/api/engine"
	"github.com/COG-GTM/openSeaChest/dashboard/api/models"
)

// CampaignHandler holds the database connection used by campaign endpoints.
type CampaignHandler struct {
	DB *sql.DB
}

// NewCampaignHandler creates a CampaignHandler with the given database connection.
func NewCampaignHandler(db *sql.DB) *CampaignHandler {
	return &CampaignHandler{DB: db}
}

// CreateCampaign handles POST /api/v1/firmware/campaigns.
// It creates a firmware update campaign targeting devices matching the
// referenced policy, then enqueues those devices with status "pending".
//
// Campaign execution instructs agents to run:
//
//	openSeaChest_Firmware --downloadFW <firmware_file> -d <device_handle>
//
// and maps the resulting exit codes as documented in
// docs/man/man8/openSeaChest_Firmware.8 (see models.ExitCodeMapping).
func (h *CampaignHandler) CreateCampaign(w http.ResponseWriter, r *http.Request) {
	var input models.FirmwareCampaignInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if input.Name == "" || input.PolicyID == 0 || input.FirmwareFile == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "name, policy_id, and firmware_file are required",
		})
		return
	}

	// Verify the referenced policy exists and load it for device matching.
	var policy models.FirmwarePolicy
	err := h.DB.QueryRow(
		`SELECT id, name, model_pattern, required_firmware, severity, created_at, updated_at
		 FROM firmware_policies WHERE id = ?`, input.PolicyID,
	).Scan(&policy.ID, &policy.Name, &policy.ModelPattern, &policy.RequiredFirmware,
		&policy.Severity, &policy.CreatedAt, &policy.UpdatedAt)
	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "policy not found"})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to look up policy"})
		return
	}

	now := time.Now().UTC()
	result, err := h.DB.Exec(
		`INSERT INTO firmware_campaigns (name, policy_id, firmware_file, status, created_at)
		 VALUES (?, ?, ?, 'created', ?)`,
		input.Name, input.PolicyID, input.FirmwareFile, now,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create campaign"})
		return
	}

	campaignID, _ := result.LastInsertId()

	// Find non-compliant devices matching the policy and enqueue them.
	devices := h.findTargetDevices(policy)
	for _, d := range devices {
		_, _ = h.DB.Exec(
			`INSERT INTO campaign_devices (campaign_id, device_serial, status, updated_at)
			 VALUES (?, ?, 'pending', ?)`,
			campaignID, d.Serial, now,
		)
	}

	campaign := models.FirmwareCampaign{
		ID:           campaignID,
		Name:         input.Name,
		PolicyID:     input.PolicyID,
		FirmwareFile: input.FirmwareFile,
		Status:       "created",
		CreatedAt:    now,
	}

	writeJSON(w, http.StatusCreated, campaign)
}

// GetCampaign handles GET /api/v1/firmware/campaigns/{id}.
// It returns the campaign details, per-device statuses, and an aggregate summary.
func (h *CampaignHandler) GetCampaign(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid campaign id"})
		return
	}

	var campaign models.FirmwareCampaign
	err = h.DB.QueryRow(
		`SELECT id, name, policy_id, firmware_file, status, created_at, started_at, completed_at
		 FROM firmware_campaigns WHERE id = ?`, id,
	).Scan(&campaign.ID, &campaign.Name, &campaign.PolicyID, &campaign.FirmwareFile,
		&campaign.Status, &campaign.CreatedAt, &campaign.StartedAt, &campaign.CompletedAt)
	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "campaign not found"})
		return
	} else if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch campaign"})
		return
	}

	rows, err := h.DB.Query(
		`SELECT id, campaign_id, device_serial, status, updated_at
		 FROM campaign_devices WHERE campaign_id = ?`, id,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch campaign devices"})
		return
	}
	defer rows.Close()

	devices := make([]models.CampaignDevice, 0)
	for rows.Next() {
		var d models.CampaignDevice
		if err := rows.Scan(&d.ID, &d.CampaignID, &d.DeviceSerial, &d.Status, &d.UpdatedAt); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to scan device row"})
			return
		}
		devices = append(devices, d)
	}

	summary := engine.ComputeCampaignSummary(devices)

	writeJSON(w, http.StatusOK, models.CampaignStatus{
		Campaign: campaign,
		Devices:  devices,
		Summary:  summary,
	})
}

// ReportDeviceResult handles POST /api/v1/firmware/campaigns/{id}/results.
// Agents call this endpoint to report the exit code from running
// openSeaChest_Firmware --downloadFW on a specific device.
func (h *CampaignHandler) ReportDeviceResult(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	campaignID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid campaign id"})
		return
	}

	var body struct {
		DeviceSerial string `json:"device_serial"`
		ExitCode     int    `json:"exit_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	status := engine.MapExitCode(body.ExitCode)
	now := time.Now().UTC()

	res, err := h.DB.Exec(
		`UPDATE campaign_devices SET status=?, updated_at=?
		 WHERE campaign_id=? AND device_serial=?`,
		status, now, campaignID, body.DeviceSerial,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update device status"})
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "device not found in campaign"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"device_serial": body.DeviceSerial,
		"status":        status,
		"exit_code":     strconv.Itoa(body.ExitCode),
	})
}

// findTargetDevices returns devices matching the policy's model pattern that
// are NOT already running the required firmware (i.e. non-compliant devices).
func (h *CampaignHandler) findTargetDevices(policy models.FirmwarePolicy) []models.Device {
	rows, err := h.DB.Query(
		`SELECT id, serial, product_identification, product_revision FROM devices`,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var targets []models.Device
	for rows.Next() {
		var d models.Device
		if err := rows.Scan(&d.ID, &d.Serial, &d.ProductIdentification, &d.ProductRevision); err != nil {
			continue
		}
		if engine.MatchModel(policy.ModelPattern, d.ProductIdentification) &&
			!engine.MatchFirmware(policy.RequiredFirmware, d.ProductRevision) {
			targets = append(targets, d)
		}
	}
	return targets
}
