// Package agent provides health data collection by wrapping openSeaChest CLI tools.
package agent

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/COG-GTM/openSeaChest/dashboard/models"
)

// openSeaChest_SMART --smartCheck exit codes mapped from eUtilExitCodes in
// include/openseachest_util_options.h:
//
//	UTIL_EXIT_NO_ERROR              = 0  -> healthy
//	UTIL_EXIT_ERROR_IN_COMMAND_LINE = 1  -> unknown (bad invocation)
//	UTIL_EXIT_OPERATION_FAILURE     = 3  -> critical (SMART test failed)
//	UTIL_EXIT_OPERATION_NOT_SUPPORTED = 4 -> unknown
//	UTIL_EXIT_OPERATION_ABORTED     = 5  -> warning (in-progress / aborted)
const (
	exitNoError             = 0
	exitErrorInCommandLine  = 1
	exitOperationFailure    = 3
	exitOperationNotSupported = 4
	exitOperationAborted    = 5
)

// smartCheckBinary is the name of the openSeaChest SMART check binary.
// It can be overridden for testing.
var smartCheckBinary = "openSeaChest_SMART"

// RunSMARTCheck executes openSeaChest_SMART --smartCheck against the specified
// device handle (e.g., /dev/sg0) and maps the exit code to a health status.
//
// Exit code mapping:
//   - 0 (UTIL_EXIT_NO_ERROR): healthy — SMART self-test passed
//   - 5 (UTIL_EXIT_OPERATION_ABORTED): warning — test in progress or aborted
//   - 3 (UTIL_EXIT_OPERATION_FAILURE): critical — SMART test reports failure
//   - Any other code: unknown
func RunSMARTCheck(deviceHandle string) (*models.SMARTCheckResult, error) {
	if deviceHandle == "" {
		return nil, fmt.Errorf("device handle must not be empty")
	}

	cmd := exec.Command(smartCheckBinary, "-d", deviceHandle, "--smartCheck")
	output, err := cmd.CombinedOutput()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to execute %s: %w", smartCheckBinary, err)
		}
	}

	status := mapExitCodeToStatus(exitCode)

	result := &models.SMARTCheckResult{
		ExitCode:  exitCode,
		Status:    status,
		RawOutput: strings.TrimSpace(string(output)),
		Timestamp: time.Now().UTC(),
	}

	return result, nil
}

// mapExitCodeToStatus converts an openSeaChest exit code to a HealthStatus.
func mapExitCodeToStatus(exitCode int) models.HealthStatus {
	switch exitCode {
	case exitNoError:
		return models.HealthStatusHealthy
	case exitOperationAborted:
		return models.HealthStatusWarning
	case exitOperationFailure:
		return models.HealthStatusCritical
	default:
		return models.HealthStatusUnknown
	}
}

// SMARTCheckToMetrics converts a SMARTCheckResult into HealthMetric entries
// suitable for storage in TimescaleDB.
func SMARTCheckToMetrics(result *models.SMARTCheckResult) []models.HealthMetric {
	statusValue := 0.0
	switch result.Status {
	case models.HealthStatusHealthy:
		statusValue = 0.0
	case models.HealthStatusWarning:
		statusValue = 1.0
	case models.HealthStatusCritical:
		statusValue = 2.0
	default:
		statusValue = -1.0
	}

	return []models.HealthMetric{
		{
			DeviceSerial: result.DeviceSerial,
			Timestamp:    result.Timestamp,
			MetricName:   "smart_check_status",
			MetricValue:  statusValue,
			Unit:         "status",
			Source:        "smart_check",
		},
		{
			DeviceSerial: result.DeviceSerial,
			Timestamp:    result.Timestamp,
			MetricName:   "smart_check_exit_code",
			MetricValue:  float64(result.ExitCode),
			Unit:         "exit_code",
			Source:        "smart_check",
		},
	}
}
