// Package engine implements the compliance evaluation logic and campaign
// execution for the firmware compliance auditing system.
package engine

import (
	"path/filepath"
	"strings"

	"github.com/COG-GTM/openSeaChest/dashboard/api/models"
)

// MatchModel returns true if the device's product_identification matches the
// policy's model_pattern using glob matching. This mirrors the MODEL_MATCH_FLAG
// logic in utils/C/openSeaChest/openSeaChest_NVMe.c where --modelMatch provides
// a closest/prefix match against the device model string.
//
// The Go filepath.Match function implements glob semantics:
//   - '*' matches any sequence of non-separator characters
//   - '?' matches any single non-separator character
//   - '[...]' matches character ranges
//
// Both the pattern and the model string are compared case-insensitively
// to align with how storage devices report model strings.
func MatchModel(modelPattern, productIdentification string) bool {
	pattern := strings.ToUpper(strings.TrimSpace(modelPattern))
	model := strings.ToUpper(strings.TrimSpace(productIdentification))

	matched, err := filepath.Match(pattern, model)
	if err != nil {
		return false
	}
	return matched
}

// MatchFirmware returns true if the device's current firmware revision matches
// the policy's required firmware exactly. This mirrors the FW_MATCH_FLAG logic
// in utils/C/openSeaChest/openSeaChest_NVMe.c where --onlyFW performs an exact
// string match on the firmware revision.
func MatchFirmware(requiredFirmware, productRevision string) bool {
	return strings.TrimSpace(requiredFirmware) == strings.TrimSpace(productRevision)
}

// EvaluateCompliance checks a single device against a single policy and returns
// a ComplianceResult. A device is compliant when:
//  1. Its product_identification matches the policy model_pattern (glob), AND
//  2. Its product_revision equals the policy required_firmware (exact match).
//
// If the model does not match, the result is nil (policy does not apply).
func EvaluateCompliance(policy models.FirmwarePolicy, device models.Device) *models.ComplianceResult {
	if !MatchModel(policy.ModelPattern, device.ProductIdentification) {
		return nil // policy does not apply to this device
	}

	compliant := MatchFirmware(policy.RequiredFirmware, device.ProductRevision)
	return &models.ComplianceResult{
		DeviceID:        device.ID,
		PolicyID:        policy.ID,
		Compliant:       compliant,
		CurrentFirmware: device.ProductRevision,
	}
}

// EvaluateFleet evaluates all devices against all policies and returns
// a FleetComplianceReport with per-policy summary counts.
func EvaluateFleet(policies []models.FirmwarePolicy, devices []models.Device) []models.PolicyComplianceSummary {
	summaries := make([]models.PolicyComplianceSummary, 0, len(policies))

	for _, policy := range policies {
		summary := models.PolicyComplianceSummary{
			PolicyID:   policy.ID,
			PolicyName: policy.Name,
			Severity:   policy.Severity,
		}

		matchedAny := false
		for _, device := range devices {
			result := EvaluateCompliance(policy, device)
			if result == nil {
				continue // device model doesn't match the policy pattern
			}
			matchedAny = true
			summary.Total++
			if result.Compliant {
				summary.Compliant++
			} else {
				summary.NonCompliant++
			}
		}

		// Devices that could not be evaluated (e.g. no agent data) are unknown.
		// For now, unknown = 0 because we only count devices we can evaluate.
		if !matchedAny {
			summary.Unknown = 0
		}

		summaries = append(summaries, summary)
	}

	return summaries
}
