package agent

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/COG-GTM/openSeaChest/dashboard/models"
)

// farmLogBinary is the name of the openSeaChest Logs binary.
var farmLogBinary = "openSeaChest_Logs"

// CollectFARMLogs executes openSeaChest_Logs --farm --logMode pipe against
// the specified device handle and parses the JSON output into FARMLogData.
//
// The --logMode pipe flag causes the tool to output JSON to stdout in the
// format documented in example/fromPipe.json.
func CollectFARMLogs(deviceHandle string) (*models.FARMLogData, error) {
	if deviceHandle == "" {
		return nil, fmt.Errorf("device handle must not be empty")
	}

	cmd := exec.Command(farmLogBinary, "-d", deviceHandle, "--farm", "--logMode", "pipe")
	output, err := cmd.Output()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return nil, fmt.Errorf("failed to execute %s: %w", farmLogBinary, err)
		}
	}

	return ParseFARMLogJSON(output)
}

// FARMLogToMetrics converts parsed FARM log data into HealthMetric entries
// for storage in TimescaleDB.
func FARMLogToMetrics(data *models.FARMLogData) []models.HealthMetric {
	now := time.Now().UTC()
	serial := data.SerialNumber

	return []models.HealthMetric{
		{
			DeviceSerial: serial,
			Timestamp:    now,
			MetricName:   "farm_power_on_hours",
			MetricValue:  data.PowerOnHours,
			Unit:         "hours",
			Source:       "farm_log",
		},
		{
			DeviceSerial: serial,
			Timestamp:    now,
			MetricName:   "farm_current_temperature",
			MetricValue:  data.CurrentTemperature,
			Unit:         "celsius",
			Source:       "farm_log",
		},
		{
			DeviceSerial: serial,
			Timestamp:    now,
			MetricName:   "farm_highest_temperature",
			MetricValue:  data.HighestTemperature,
			Unit:         "celsius",
			Source:       "farm_log",
		},
		{
			DeviceSerial: serial,
			Timestamp:    now,
			MetricName:   "farm_lowest_temperature",
			MetricValue:  data.LowestTemperature,
			Unit:         "celsius",
			Source:       "farm_log",
		},
		{
			DeviceSerial: serial,
			Timestamp:    now,
			MetricName:   "farm_unrecoverable_read_errors",
			MetricValue:  data.UnrecoverableReadErrors,
			Unit:         "count",
			Source:       "farm_log",
		},
		{
			DeviceSerial: serial,
			Timestamp:    now,
			MetricName:   "farm_unrecoverable_write_errors",
			MetricValue:  data.UnrecoverableWriteErrors,
			Unit:         "count",
			Source:       "farm_log",
		},
		{
			DeviceSerial: serial,
			Timestamp:    now,
			MetricName:   "farm_rated_workload_percentage",
			MetricValue:  data.RatedWorkloadPercentage,
			Unit:         "percent",
			Source:       "farm_log",
		},
		{
			DeviceSerial: serial,
			Timestamp:    now,
			MetricName:   "farm_helium_pressure_threshold_tripped",
			MetricValue:  data.HeliumPressureThresholdTripped,
			Unit:         "boolean",
			Source:       "farm_log",
		},
		{
			DeviceSerial: serial,
			Timestamp:    now,
			MetricName:   "farm_power_cycle_count",
			MetricValue:  data.PowerCycleCount,
			Unit:         "count",
			Source:       "farm_log",
		},
		{
			DeviceSerial: serial,
			Timestamp:    now,
			MetricName:   "farm_head_load_events",
			MetricValue:  data.HeadLoadEvents,
			Unit:         "count",
			Source:       "farm_log",
		},
		{
			DeviceSerial: serial,
			Timestamp:    now,
			MetricName:   "farm_mechanical_start_failures",
			MetricValue:  data.MechanicalStartFailures,
			Unit:         "count",
			Source:       "farm_log",
		},
		{
			DeviceSerial: serial,
			Timestamp:    now,
			MetricName:   "farm_reallocated_sector_reclamations",
			MetricValue:  data.ReallocatedSectorReclamations,
			Unit:         "count",
			Source:       "farm_log",
		},
	}
}

// farmRawJSON mirrors the top-level structure of the FARM log JSON output
// from openSeaChest_Logs --farm --logMode pipe. See example/fromPipe.json.
type farmRawJSON struct {
	DriveInfo   map[string]interface{} `json:"Drive Information From Farm Log copy 0"`
	Workload    map[string]interface{} `json:"Workload From Farm Log copy 0"`
	ErrorInfo   map[string]interface{} `json:"Error Information Log From Farm Log copy 0"`
	EnvInfo     map[string]interface{} `json:"Environment Information From Farm Log copy 0"`
	Reliability map[string]interface{} `json:"Reliability Information From Farm Log copy 0"`
	Actuator    map[string]interface{} `json:"LUN Actuator Information 0x0 From Farm Log copy 0"`
}

// ParseFARMLogJSON parses the raw JSON bytes from openSeaChest_Logs --farm --logMode pipe
// and extracts key reliability metrics into a FARMLogData struct.
//
// The JSON structure is documented in example/fromPipe.json. Key fields extracted:
//   - Serial Number, Model Number, Firmware Rev (from Drive Information)
//   - Power on Hour, Power Cycle count (from Drive Information)
//   - Current Temperature (from Environment Information)
//   - Unrecoverable Read Errors, Unrecoverable Write Errors (from Error Information)
//   - Rated Workload Percentage (from Workload)
//   - Helium Pressure Threshold Tripped (from Reliability Information)
//   - Head Load Events, Reallocated Sector Reclamations (from Actuator Information)
func ParseFARMLogJSON(data []byte) (*models.FARMLogData, error) {
	var raw farmRawJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse FARM log JSON: %w", err)
	}

	result := &models.FARMLogData{}

	// Extract Drive Information fields
	if raw.DriveInfo != nil {
		result.SerialNumber = extractString(raw.DriveInfo, "Serial Number")
		result.ModelNumber = extractString(raw.DriveInfo, "Model Number")
		result.FirmwareRev = extractString(raw.DriveInfo, "Firmware Rev")
		result.PowerOnHours = extractFloat(raw.DriveInfo, "Power on Hour")
		result.PowerCycleCount = extractFloat(raw.DriveInfo, "Power Cycle count")
	}

	// Extract Workload fields
	if raw.Workload != nil {
		result.RatedWorkloadPercentage = extractFloat(raw.Workload, "Rated Workload Percentage")
	}

	// Extract Error Information fields
	if raw.ErrorInfo != nil {
		result.UnrecoverableReadErrors = extractFloat(raw.ErrorInfo, "Unrecoverable Read Errors")
		result.UnrecoverableWriteErrors = extractFloat(raw.ErrorInfo, "Unrecoverable Write Errors")
		result.MechanicalStartFailures = extractFloat(raw.ErrorInfo, "Number of Mechanical Start Failures")
	}

	// Extract Environment Information fields.
	// FARM JSON quirk: successive field values in Environment Information are
	// concatenated into a single string. Each field's string = previous field's
	// string + current value appended. For example:
	//   Current Temperature:  "...copy 037"       -> value = 37
	//   Highest Temperature:  "...copy 03756"     -> value = 56  (strip prev "...037")
	//   Lowest Temperature:   "...copy 0375624"   -> value = 24  (strip prev "...03756")
	// We use extractChainedFloats to do differential parsing.
	if raw.EnvInfo != nil {
		tempFields := []string{
			"Current Temperature (Celsius)",
			"Highest Temperature",
			"Lowest Temperature",
		}
		tempValues := extractChainedFloats(raw.EnvInfo, tempFields)
		if len(tempValues) > 0 {
			result.CurrentTemperature = tempValues[0]
		}
		if len(tempValues) > 1 {
			result.HighestTemperature = tempValues[1]
		}
		if len(tempValues) > 2 {
			result.LowestTemperature = tempValues[2]
		}
	}

	// Extract Reliability Information fields
	if raw.Reliability != nil {
		result.HeliumPressureThresholdTripped = extractFloat(raw.Reliability, "Helium Pressure Threshold Tripped")
	}

	// Extract Actuator Information fields
	if raw.Actuator != nil {
		result.HeadLoadEvents = extractFloat(raw.Actuator, "Head Load Events")
		result.ReallocatedSectorReclamations = extractFloat(raw.Actuator, "Number of Reallocated Sector Reclamations")
	}

	return result, nil
}

// extractFloat safely extracts a numeric value from a map[string]interface{}.
// Handles both direct numeric types and string representations.
func extractFloat(m map[string]interface{}, key string) float64 {
	val, ok := m[key]
	if !ok {
		return 0
	}

	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case json.Number:
		f, err := v.Float64()
		if err != nil {
			return 0
		}
		return f
	case string:
		// Some fields in the FARM JSON contain string representations of numbers.
		return parseNumericString(v)
	default:
		return 0
	}
}

// extractString safely extracts a string value from a map[string]interface{}.
func extractString(m map[string]interface{}, key string) string {
	val, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := val.(string)
	if !ok {
		return fmt.Sprintf("%v", val)
	}
	return s
}

// extractChainedFloats handles the FARM JSON concatenation quirk where successive
// field values in a section are appended into a growing string. Each field's raw
// string = previous field's raw string + current value appended.
//
// For example, with fields ["Current Temperature", "Highest Temperature", "Lowest Temperature"]:
//   Field 0: "Environment Information From Farm Log copy 037"       -> strip prefix -> "37"
//   Field 1: "Environment Information From Farm Log copy 03756"     -> strip field 0 string -> "56"
//   Field 2: "Environment Information From Farm Log copy 0375624"   -> strip field 1 string -> "24"
//
// If a field has a direct numeric value (not a string), it is used as-is.
func extractChainedFloats(m map[string]interface{}, keys []string) []float64 {
	results := make([]float64, len(keys))
	prevRaw := ""

	for i, key := range keys {
		val, ok := m[key]
		if !ok {
			continue
		}

		switch v := val.(type) {
		case float64:
			results[i] = v
			// For numeric values, prevRaw stays as-is (no string to chain from)
		case string:
			// Strip the previous field's raw string to isolate this field's value.
			suffix := v
			if prevRaw != "" && len(v) > len(prevRaw) {
				suffix = v[len(prevRaw):]
			}
			results[i] = parseNumericString(suffix)
			prevRaw = v
		default:
			// Unsupported type, leave as 0
		}
	}

	return results
}

// parseNumericString parses a float from a string, ignoring leading non-numeric
// characters. Falls back to 0 if no number can be extracted.
func parseNumericString(s string) float64 {
	numStr := ""
	foundDigit := false
	for _, ch := range s {
		if ch >= '0' && ch <= '9' || ch == '.' || (ch == '-' && !foundDigit) {
			numStr += string(ch)
			foundDigit = true
		} else if foundDigit {
			break
		}
	}

	if numStr == "" {
		return 0
	}

	var f float64
	_, err := fmt.Sscanf(numStr, "%f", &f)
	if err != nil {
		return 0
	}
	return f
}
