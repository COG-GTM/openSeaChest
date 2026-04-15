// Package services implements the core business logic for the alerting engine.
package services

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/COG-GTM/openSeaChest/dashboard/api/models"
)

// DeviceReport represents telemetry data reported by an agent for a device.
type DeviceReport struct {
	DeviceSerial          string    `json:"device_serial"`
	SMARTStatus           string    `json:"smart_status"`            // "passed", "failed", "warning", "in_progress"
	TemperatureCelsius    float64   `json:"temperature_celsius"`
	CurrentFirmware       string    `json:"current_firmware"`
	LastReportedAt        time.Time `json:"last_reported_at"`
	ReallocatedSectorCount int      `json:"reallocated_sector_count"`
	PrevReallocatedCount  int       `json:"prev_reallocated_count"`
}

// RuleEngine evaluates device reports against configured alert rules.
type RuleEngine struct {
	store    *Store
	notifier *NotificationService
}

// NewRuleEngine creates a new RuleEngine.
func NewRuleEngine(store *Store, notifier *NotificationService) *RuleEngine {
	return &RuleEngine{
		store:    store,
		notifier: notifier,
	}
}

// Evaluate checks a device report against all enabled rules.
func (re *RuleEngine) Evaluate(report DeviceReport) {
	rules := re.store.ListRules()
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		re.evaluateRule(rule, report)
	}
}

// evaluateRule checks a single rule against a device report.
func (re *RuleEngine) evaluateRule(rule models.AlertRule, report DeviceReport) {
	var shouldFire bool
	var message string

	switch rule.RuleType {
	case models.RuleTypeSMARTTrip:
		shouldFire, message = re.evaluateSMARTTrip(report)
	case models.RuleTypeSMARTWarning:
		shouldFire, message = re.evaluateSMARTWarning(report)
	case models.RuleTypeTemperatureThreshold:
		shouldFire, message = re.evaluateTemperatureThreshold(rule, report)
		// Temperature has special debounce logic handled inside the evaluate function
		if !shouldFire {
			return
		}
	case models.RuleTypeFirmwareNonCompliance:
		shouldFire, message = re.evaluateFirmwareNonCompliance(rule, report)
	case models.RuleTypeAgentStale:
		shouldFire, message = re.evaluateAgentStale(rule, report)
	case models.RuleTypeReallocatedSectorGrowth:
		shouldFire, message = re.evaluateReallocatedSectorGrowth(rule, report)
	default:
		return
	}

	if shouldFire {
		alert, isNew, err := re.store.CreateOrDeduplicateAlert(rule.ID, report.DeviceSerial, message)
		if err != nil {
			log.Printf("ERROR: failed to create/deduplicate alert for rule %s: %v", rule.ID, err)
			return
		}

		if isNew {
			re.sendNotifications(rule, alert)
		}
	}
}

// evaluateSMARTTrip checks if the SMART self-test has failed.
// Maps to openSeaChest exit code UTIL_EXIT_OPERATION_FAILURE (3) from SMART tests.
func (re *RuleEngine) evaluateSMARTTrip(report DeviceReport) (bool, string) {
	if report.SMARTStatus == "failed" {
		return true, fmt.Sprintf("SMART self-test failure detected on device %s", report.DeviceSerial)
	}
	return false, ""
}

// evaluateSMARTWarning checks if SMART status is warning or in-progress.
// Maps to openSeaChest exit codes where SMART check returns non-pass status.
func (re *RuleEngine) evaluateSMARTWarning(report DeviceReport) (bool, string) {
	if report.SMARTStatus == "warning" || report.SMARTStatus == "in_progress" {
		return true, fmt.Sprintf("SMART check returned %s status on device %s",
			report.SMARTStatus, report.DeviceSerial)
	}
	return false, ""
}

// evaluateTemperatureThreshold checks if device temperature exceeds the configured
// threshold. Implements debounce: only fires after 2 consecutive intervals above
// the threshold to prevent transient spike noise.
func (re *RuleEngine) evaluateTemperatureThreshold(rule models.AlertRule, report DeviceReport) (bool, string) {
	var config models.TemperatureThresholdConfig
	if err := json.Unmarshal(rule.Config, &config); err != nil {
		log.Printf("ERROR: invalid temperature threshold config for rule %s: %v", rule.ID, err)
		return false, ""
	}

	consecutiveRequired := config.ConsecutiveCount
	if consecutiveRequired <= 0 {
		consecutiveRequired = 2 // default debounce: 2 consecutive intervals
	}

	if report.TemperatureCelsius > config.ThresholdCelsius {
		count := re.store.IncrementTempCount(rule.ID, report.DeviceSerial)
		if count >= consecutiveRequired {
			return true, fmt.Sprintf("Temperature %.1f°C exceeds threshold %.1f°C on device %s (%d consecutive readings)",
				report.TemperatureCelsius, config.ThresholdCelsius, report.DeviceSerial, count)
		}
		// Not enough consecutive readings yet
		return false, ""
	}

	// Temperature is below threshold, reset the counter
	re.store.ResetTempCount(rule.ID, report.DeviceSerial)
	return false, ""
}

// evaluateFirmwareNonCompliance checks if device firmware doesn't match the required policy.
func (re *RuleEngine) evaluateFirmwareNonCompliance(rule models.AlertRule, report DeviceReport) (bool, string) {
	var config models.FirmwareNonComplianceConfig
	if err := json.Unmarshal(rule.Config, &config); err != nil {
		log.Printf("ERROR: invalid firmware non-compliance config for rule %s: %v", rule.ID, err)
		return false, ""
	}

	if config.RequiredFirmware != "" && report.CurrentFirmware != config.RequiredFirmware {
		return true, fmt.Sprintf("Device %s firmware %q does not match required firmware %q",
			report.DeviceSerial, report.CurrentFirmware, config.RequiredFirmware)
	}
	return false, ""
}

// evaluateAgentStale checks if the agent hasn't reported within the configured interval.
func (re *RuleEngine) evaluateAgentStale(rule models.AlertRule, report DeviceReport) (bool, string) {
	var config models.AgentStaleConfig
	if err := json.Unmarshal(rule.Config, &config); err != nil {
		log.Printf("ERROR: invalid agent stale config for rule %s: %v", rule.ID, err)
		return false, ""
	}

	if config.StaleIntervalSeconds <= 0 {
		return false, ""
	}

	staleDuration := time.Duration(config.StaleIntervalSeconds) * time.Second
	if report.LastReportedAt.IsZero() {
		// No timestamp provided; skip to avoid false positives
		return false, ""
	}
	if time.Since(report.LastReportedAt) > staleDuration {
		return true, fmt.Sprintf("Agent for device %s has not reported for %s (threshold: %s)",
			report.DeviceSerial,
			time.Since(report.LastReportedAt).Round(time.Second),
			staleDuration)
	}
	return false, ""
}

// evaluateReallocatedSectorGrowth checks if the reallocated sector count is increasing.
func (re *RuleEngine) evaluateReallocatedSectorGrowth(rule models.AlertRule, report DeviceReport) (bool, string) {
	var config models.ReallocatedSectorGrowthConfig
	if err := json.Unmarshal(rule.Config, &config); err != nil {
		log.Printf("ERROR: invalid reallocated sector growth config for rule %s: %v", rule.ID, err)
		return false, ""
	}

	threshold := config.GrowthThreshold
	if threshold <= 0 {
		threshold = 1 // default: any growth triggers alert
	}

	growth := report.ReallocatedSectorCount - report.PrevReallocatedCount
	if growth >= threshold {
		return true, fmt.Sprintf("Reallocated sector count grew by %d on device %s (current: %d, previous: %d)",
			growth, report.DeviceSerial, report.ReallocatedSectorCount, report.PrevReallocatedCount)
	}
	return false, ""
}

// sendNotifications dispatches notifications through configured channels for a rule.
func (re *RuleEngine) sendNotifications(rule models.AlertRule, alert *models.Alert) {
	if re.notifier == nil {
		return
	}

	payload := models.NotificationPayload{
		AlertID:      alert.ID,
		RuleName:     rule.Name,
		RuleType:     rule.RuleType,
		DeviceSerial: alert.DeviceSerial,
		Severity:     rule.Severity,
		State:        alert.State,
		Message:      alert.Message,
		FirstSeen:    alert.FirstSeen,
		LastSeen:     alert.LastSeen,
		Timestamp:    time.Now().UTC(),
	}

	for _, channelID := range rule.NotificationChannels {
		ch, err := re.store.GetChannel(channelID)
		if err != nil {
			log.Printf("ERROR: notification channel %s not found: %v", channelID, err)
			continue
		}
		if !ch.Enabled {
			continue
		}

		go func(channel models.NotificationChannel) {
			if err := re.notifier.Send(channel, payload); err != nil {
				log.Printf("ERROR: failed to send notification via channel %s (%s): %v",
					channel.ID, channel.Type, err)
			}
		}(*ch)
	}
}
