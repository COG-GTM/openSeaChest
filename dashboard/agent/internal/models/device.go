// Package models defines data structures shared across the agent.
package models

// Device represents a storage device discovered by openSeaChest.
type Device struct {
	DeviceHandle  string `json:"device_handle"`            // e.g. /dev/sg0
	Model         string `json:"model"`                    // product_identification
	SerialNumber  string `json:"serial_number"`            // drive serial
	FirmwareRev   string `json:"firmware_rev"`             // product_revision
	CapacityBytes int64  `json:"capacity_bytes,omitempty"` // raw capacity in bytes
	InterfaceType string `json:"interface_type"`           // SATA, SAS, NVMe, USB
	WWN           string `json:"wwn,omitempty"`            // World Wide Name
}

// InventoryReport is the payload the agent POSTs to the central API.
type InventoryReport struct {
	AgentID  string   `json:"agent_id"`
	Hostname string   `json:"hostname"`
	Devices  []Device `json:"devices"`
}
