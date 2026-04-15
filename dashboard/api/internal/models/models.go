// Package models defines request/response types for the Fleet Inventory API.
package models

import "time"

// --- Request models ---

// DeviceReport is a single device entry inside an agent inventory report.
type DeviceReport struct {
	DeviceHandle  string `json:"device_handle"`
	Model         string `json:"model"`
	SerialNumber  string `json:"serial_number"`
	FirmwareRev   string `json:"firmware_rev"`
	CapacityBytes int64  `json:"capacity_bytes,omitempty"`
	InterfaceType string `json:"interface_type"`
	WWN           string `json:"wwn,omitempty"`
}

// InventoryReportRequest is the JSON body POSTed by agents.
type InventoryReportRequest struct {
	AgentID  string         `json:"agent_id"`
	Hostname string         `json:"hostname"`
	Devices  []DeviceReport `json:"devices"`
}

// --- Response models ---

// DeviceResponse is a single device returned by the API.
type DeviceResponse struct {
	ID            int64     `json:"id"`
	HostID        int64     `json:"host_id"`
	Hostname      string    `json:"hostname,omitempty"`
	SerialNumber  string    `json:"serial_number"`
	Model         string    `json:"model"`
	FirmwareRev   string    `json:"firmware_rev"`
	CapacityBytes int64     `json:"capacity_bytes"`
	InterfaceType string    `json:"interface_type"`
	WWN           string    `json:"wwn,omitempty"`
	FirstSeen     time.Time `json:"first_seen"`
	LastSeen      time.Time `json:"last_seen"`
}

// HostResponse represents a registered host/agent.
type HostResponse struct {
	ID       int64     `json:"id"`
	AgentID  string    `json:"agent_id"`
	Hostname string    `json:"hostname"`
	LastSeen time.Time `json:"last_seen"`
}

// PaginatedDevices wraps a page of device results.
type PaginatedDevices struct {
	Devices []DeviceResponse `json:"devices"`
	Total   int              `json:"total"`
	Page    int              `json:"page"`
	PerPage int              `json:"per_page"`
}

// ErrorResponse is the standard error envelope.
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}
