// Package models defines the data structures used across the firmware
// compliance auditing system. These types map directly to the database
// schema defined in migrations/001_firmware_compliance.sql.
package models

import "time"

// FirmwarePolicy describes a firmware requirement for a set of devices
// identified by a glob pattern on the model (product_identification) field.
type FirmwarePolicy struct {
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	ModelPattern     string    `json:"model_pattern"`      // glob pattern, e.g. "ST500LM*"
	RequiredFirmware string    `json:"required_firmware"`   // exact firmware revision required
	Severity         string    `json:"severity"`            // "critical", "warning", "info"
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// FirmwarePolicyInput is the request body for creating or updating a policy.
type FirmwarePolicyInput struct {
	Name             string `json:"name"`
	ModelPattern     string `json:"model_pattern"`
	RequiredFirmware string `json:"required_firmware"`
	Severity         string `json:"severity"`
}

// ComplianceResult stores the outcome of evaluating a single device against a
// single policy.
type ComplianceResult struct {
	ID              int64     `json:"id"`
	DeviceID        string    `json:"device_id"`
	PolicyID        int64     `json:"policy_id"`
	Compliant       bool      `json:"compliant"`
	CurrentFirmware string    `json:"current_firmware"`
	CheckedAt       time.Time `json:"checked_at"`
}

// Device represents a storage device in the fleet as reported by agents.
type Device struct {
	ID                    string `json:"id"`
	Serial                string `json:"serial"`
	ProductIdentification string `json:"product_identification"` // model string
	ProductRevision       string `json:"product_revision"`       // current firmware revision
}

// PolicyComplianceSummary is one entry in the fleet compliance report.
type PolicyComplianceSummary struct {
	PolicyID     int64  `json:"policy_id"`
	PolicyName   string `json:"policy_name"`
	Severity     string `json:"severity"`
	Compliant    int    `json:"compliant"`
	NonCompliant int    `json:"non_compliant"`
	Unknown      int    `json:"unknown"`
	Total        int    `json:"total"`
}

// FleetComplianceReport is the response body for GET /api/v1/firmware/compliance.
type FleetComplianceReport struct {
	GeneratedAt time.Time                 `json:"generated_at"`
	Policies    []PolicyComplianceSummary `json:"policies"`
}

// FirmwareCampaign tracks a firmware update rollout targeting devices that
// match a specific policy.
type FirmwareCampaign struct {
	ID           int64      `json:"id"`
	Name         string     `json:"name"`
	PolicyID     int64      `json:"policy_id"`
	FirmwareFile string     `json:"firmware_file"`
	Status       string     `json:"status"` // "created", "in_progress", "completed", "cancelled"
	CreatedAt    time.Time  `json:"created_at"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

// FirmwareCampaignInput is the request body for creating a campaign.
type FirmwareCampaignInput struct {
	Name         string `json:"name"`
	PolicyID     int64  `json:"policy_id"`
	FirmwareFile string `json:"firmware_file"`
}

// CampaignDevice holds per-device status within a campaign.
type CampaignDevice struct {
	ID           int64     `json:"id"`
	CampaignID   int64     `json:"campaign_id"`
	DeviceSerial string    `json:"device_serial"`
	Status       string    `json:"status"` // "pending","in_progress","success","deferred","skipped","wrong_fw","failed"
	UpdatedAt    time.Time `json:"updated_at"`
}

// CampaignStatus is the response body for GET /api/v1/firmware/campaigns/{id}.
type CampaignStatus struct {
	Campaign FirmwareCampaign `json:"campaign"`
	Devices  []CampaignDevice `json:"devices"`
	Summary  CampaignSummary  `json:"summary"`
}

// CampaignSummary provides aggregate counts for a campaign.
type CampaignSummary struct {
	Total    int `json:"total"`
	Pending  int `json:"pending"`
	Success  int `json:"success"`
	Deferred int `json:"deferred"`
	Skipped  int `json:"skipped"`
	WrongFW  int `json:"wrong_fw"`
	Failed   int `json:"failed"`
}

// ExitCodeMapping maps openSeaChest_Firmware exit codes to campaign device
// statuses. Reference: docs/man/man8/openSeaChest_Firmware.8
//
// Exit codes (from eUtilExitCodes and firmware-specific codes starting at 32):
//
//	0  (UTIL_EXIT_NO_ERROR)           → success
//	32 (Firmware Download Complete)    → success
//	33 (Deferred FW Download Complete) → deferred (reboot required)
//	38 (Firmware Already up to date)   → skipped
//	36 (Model matched, FW mismatched)  → wrong_fw
//	3  (UTIL_EXIT_OPERATION_FAILURE)   → failed
var ExitCodeMapping = map[int]string{
	0:  "success",
	32: "success",
	33: "deferred",
	38: "skipped",
	36: "wrong_fw",
	3:  "failed",
}
