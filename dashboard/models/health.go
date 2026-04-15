// Package models defines data structures for the openSeaChest Fleet Health Dashboard.
package models

import "time"

// HealthStatus represents the overall health state of a drive.
type HealthStatus string

const (
	HealthStatusHealthy  HealthStatus = "healthy"
	HealthStatusWarning  HealthStatus = "warning"
	HealthStatusCritical HealthStatus = "critical"
	HealthStatusUnknown  HealthStatus = "unknown"
)

// HealthMetric represents a single time-series health data point stored in TimescaleDB.
type HealthMetric struct {
	DeviceSerial string    `json:"device_serial"`
	Timestamp    time.Time `json:"timestamp"`
	MetricName   string    `json:"metric_name"`
	MetricValue  float64   `json:"metric_value"`
	Unit         string    `json:"unit,omitempty"`
	Source       string    `json:"source"` // "smart_check", "smart_attributes", "farm_log"
}

// DeviceHealthSummary is the current health overview for a single device.
type DeviceHealthSummary struct {
	DeviceSerial string       `json:"device_serial"`
	Status       HealthStatus `json:"status"`
	LastChecked  time.Time    `json:"last_checked"`
	Metrics      []HealthMetric `json:"metrics,omitempty"`
	Message      string       `json:"message,omitempty"`
}

// FleetHealthSummary provides aggregate health counts across all monitored devices.
type FleetHealthSummary struct {
	TotalDevices   int            `json:"total_devices"`
	HealthyCounts  int            `json:"healthy_count"`
	WarningCounts  int            `json:"warning_count"`
	CriticalCounts int            `json:"critical_count"`
	UnknownCounts  int            `json:"unknown_count"`
	Timestamp      time.Time      `json:"timestamp"`
	Devices        []DeviceHealthSummary `json:"devices,omitempty"`
}

// HealthHistoryQuery defines the parameters for a time-series health query.
type HealthHistoryQuery struct {
	DeviceSerial string    `json:"device_serial"`
	From         time.Time `json:"from"`
	To           time.Time `json:"to"`
	MetricName   string    `json:"metric,omitempty"`
}

// HealthHistoryResponse wraps the time-series data returned for a history query.
type HealthHistoryResponse struct {
	DeviceSerial string         `json:"device_serial"`
	From         time.Time      `json:"from"`
	To           time.Time      `json:"to"`
	MetricName   string         `json:"metric,omitempty"`
	DataPoints   []HealthMetric `json:"data_points"`
}

// SMARTCheckResult maps openSeaChest_SMART --smartCheck exit codes to health states.
// Exit code 0 = healthy, in-progress = warning, failure = critical.
type SMARTCheckResult struct {
	DeviceSerial string       `json:"device_serial"`
	ExitCode     int          `json:"exit_code"`
	Status       HealthStatus `json:"status"`
	RawOutput    string       `json:"raw_output,omitempty"`
	Timestamp    time.Time    `json:"timestamp"`
}

// FARMLogData represents parsed fields from openSeaChest_Logs --farm --logMode pipe JSON output.
type FARMLogData struct {
	SerialNumber                  string  `json:"serial_number"`
	ModelNumber                   string  `json:"model_number"`
	FirmwareRev                   string  `json:"firmware_rev"`
	PowerOnHours                  float64 `json:"power_on_hours"`
	CurrentTemperature            float64 `json:"current_temperature_celsius"`
	HighestTemperature            float64 `json:"highest_temperature_celsius"`
	LowestTemperature             float64 `json:"lowest_temperature_celsius"`
	UnrecoverableReadErrors       float64 `json:"unrecoverable_read_errors"`
	UnrecoverableWriteErrors      float64 `json:"unrecoverable_write_errors"`
	RatedWorkloadPercentage       float64 `json:"rated_workload_percentage"`
	HeliumPressureThresholdTripped float64 `json:"helium_pressure_threshold_tripped"`
	PowerCycleCount               float64 `json:"power_cycle_count"`
	HeadLoadEvents                float64 `json:"head_load_events"`
	MechanicalStartFailures       float64 `json:"mechanical_start_failures"`
	ReallocatedSectorReclamations float64 `json:"reallocated_sector_reclamations"`
}

// APIError is a standard error response returned by the REST API.
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
