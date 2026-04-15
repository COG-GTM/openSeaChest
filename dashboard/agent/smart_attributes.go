package agent

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/COG-GTM/openSeaChest/dashboard/models"
)

// smartAttributesBinary is the name of the SMART attributes binary.
var smartAttributesBinary = "openSeaChest_SMART"

// CollectSMARTAttributes executes openSeaChest_SMART --smartAttributes hybrid
// against the specified device handle and parses the tabular output into
// HealthMetric entries.
//
// The "hybrid" mode outputs SMART attributes in a human-readable table with
// columns: ID, Name, Flags, Value, Worst, Threshold, Raw Value.
// Each attribute row is parsed into a metric with the attribute name as the
// metric_name and the raw value as the metric_value.
func CollectSMARTAttributes(deviceHandle, deviceSerial string) ([]models.HealthMetric, error) {
	if deviceHandle == "" {
		return nil, fmt.Errorf("device handle must not be empty")
	}

	cmd := exec.Command(smartAttributesBinary, "-d", deviceHandle, "--smartAttributes", "hybrid")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return nil, fmt.Errorf("failed to execute %s: %w", smartAttributesBinary, err)
		}
	}

	return parseSMARTAttributeOutput(string(output), deviceSerial)
}

// parseSMARTAttributeOutput parses the text output of --smartAttributes hybrid.
// Lines with SMART attribute data are expected to have fields separated by
// whitespace. We look for lines that start with a numeric ID.
func parseSMARTAttributeOutput(output, deviceSerial string) ([]models.HealthMetric, error) {
	var metrics []models.HealthMetric
	now := time.Now().UTC()

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 7 {
			continue
		}

		// First field should be a numeric attribute ID.
		_, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}

		attrName := fields[1]

		// Try to parse the raw value (last field).
		rawValueStr := fields[len(fields)-1]
		rawValue, err := strconv.ParseFloat(rawValueStr, 64)
		if err != nil {
			// Some raw values may contain non-numeric characters; skip those.
			continue
		}

		// Also parse the normalized value (4th field, 0-indexed at 3).
		normalizedValue := 0.0
		if val, parseErr := strconv.ParseFloat(fields[3], 64); parseErr == nil {
			normalizedValue = val
		}

		metrics = append(metrics, models.HealthMetric{
			DeviceSerial: deviceSerial,
			Timestamp:    now,
			MetricName:   fmt.Sprintf("smart_attr_%s_raw", strings.ToLower(attrName)),
			MetricValue:  rawValue,
			Unit:         "raw",
			Source:       "smart_attributes",
		})

		metrics = append(metrics, models.HealthMetric{
			DeviceSerial: deviceSerial,
			Timestamp:    now,
			MetricName:   fmt.Sprintf("smart_attr_%s_normalized", strings.ToLower(attrName)),
			MetricValue:  normalizedValue,
			Unit:         "normalized",
			Source:       "smart_attributes",
		})
	}

	if scanErr := scanner.Err(); scanErr != nil {
		return metrics, fmt.Errorf("error scanning SMART attribute output: %w", scanErr)
	}

	return metrics, nil
}
