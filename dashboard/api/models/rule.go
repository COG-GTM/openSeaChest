// Package models defines the data structures for the alerting engine.
package models

import (
	"encoding/json"
	"time"
)

// RuleType represents a built-in alert rule type.
type RuleType string

const (
	RuleTypeSMARTTrip              RuleType = "smart_trip"
	RuleTypeSMARTWarning           RuleType = "smart_warning"
	RuleTypeTemperatureThreshold   RuleType = "temperature_threshold"
	RuleTypeFirmwareNonCompliance  RuleType = "firmware_non_compliance"
	RuleTypeAgentStale             RuleType = "agent_stale"
	RuleTypeReallocatedSectorGrowth RuleType = "reallocated_sector_growth"
)

// AllRuleTypes returns all supported built-in rule types.
func AllRuleTypes() []RuleType {
	return []RuleType{
		RuleTypeSMARTTrip,
		RuleTypeSMARTWarning,
		RuleTypeTemperatureThreshold,
		RuleTypeFirmwareNonCompliance,
		RuleTypeAgentStale,
		RuleTypeReallocatedSectorGrowth,
	}
}

// IsValidRuleType checks whether a given rule type string is valid.
func IsValidRuleType(rt RuleType) bool {
	for _, valid := range AllRuleTypes() {
		if rt == valid {
			return true
		}
	}
	return false
}

// Severity represents alert severity levels.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityWarning  Severity = "warning"
	SeverityInfo     Severity = "info"
)

// AlertRule defines a rule that triggers alerts based on device conditions.
type AlertRule struct {
	ID                   string          `json:"id"`
	Name                 string          `json:"name"`
	RuleType             RuleType        `json:"rule_type"`
	Config               json.RawMessage `json:"config"`
	Severity             Severity        `json:"severity"`
	Enabled              bool            `json:"enabled"`
	NotificationChannels []string        `json:"notification_channels,omitempty"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

// CreateRuleRequest is the request body for creating a new alert rule.
type CreateRuleRequest struct {
	Name                 string          `json:"name"`
	RuleType             RuleType        `json:"rule_type"`
	Config               json.RawMessage `json:"config"`
	Severity             Severity        `json:"severity"`
	Enabled              *bool           `json:"enabled,omitempty"`
	NotificationChannels []string        `json:"notification_channels,omitempty"`
}

// UpdateRuleRequest is the request body for updating an existing alert rule.
type UpdateRuleRequest struct {
	Name                 *string          `json:"name,omitempty"`
	Config               *json.RawMessage `json:"config,omitempty"`
	Severity             *Severity        `json:"severity,omitempty"`
	Enabled              *bool            `json:"enabled,omitempty"`
	NotificationChannels []string         `json:"notification_channels,omitempty"`
}

// RuleListResponse is the response for listing alert rules.
type RuleListResponse struct {
	Rules []AlertRule `json:"rules"`
}

// TemperatureThresholdConfig is the configuration for temperature_threshold rules.
type TemperatureThresholdConfig struct {
	ThresholdCelsius   float64 `json:"threshold_celsius"`
	ConsecutiveCount   int     `json:"consecutive_count"`   // defaults to 2 for debounce
}

// AgentStaleConfig is the configuration for agent_stale rules.
type AgentStaleConfig struct {
	StaleIntervalSeconds int `json:"stale_interval_seconds"`
}

// FirmwareNonComplianceConfig is the configuration for firmware_non_compliance rules.
type FirmwareNonComplianceConfig struct {
	RequiredFirmware string `json:"required_firmware"`
}

// ReallocatedSectorGrowthConfig is the configuration for reallocated_sector_growth rules.
type ReallocatedSectorGrowthConfig struct {
	GrowthThreshold int `json:"growth_threshold"` // minimum sector count increase to trigger
}
