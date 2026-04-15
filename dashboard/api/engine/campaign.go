package engine

import (
	"github.com/COG-GTM/openSeaChest/dashboard/api/models"
)

// MapExitCode translates an openSeaChest_Firmware exit code into a campaign
// device status string. The mapping is derived from:
//   - docs/man/man8/openSeaChest_Firmware.8 (return codes section)
//   - include/openseachest_util_options.h    (eUtilExitCodes enum)
//
// Exit code semantics:
//
//	0  → UTIL_EXIT_NO_ERROR              → "success"
//	32 → Firmware Download Complete       → "success"
//	33 → Deferred FW Download Complete    → "deferred" (reboot required)
//	38 → Firmware Already up to date      → "skipped"
//	36 → Model matched, FW mismatched     → "wrong_fw"
//	3  → UTIL_EXIT_OPERATION_FAILURE      → "failed"
//
// Any unrecognised exit code maps to "failed".
func MapExitCode(exitCode int) string {
	if status, ok := models.ExitCodeMapping[exitCode]; ok {
		return status
	}
	return "failed"
}

// BuildFirmwareCommand returns the command and arguments that an agent should
// execute on a target device. Returning a []string (suitable for exec.Command)
// avoids shell interpretation and eliminates command injection risks from
// untrusted firmware file paths or device handles.
//
// The command mirrors the usage documented in
// docs/man/man8/openSeaChest_Firmware.8.
//
// Example result:
//
//	[]string{"openSeaChest_Firmware", "--downloadFW", "/path/to/firmware.bin", "-d", "/dev/sg0"}
func BuildFirmwareCommand(firmwareFile string, deviceHandle string) []string {
	return []string{"openSeaChest_Firmware", "--downloadFW", firmwareFile, "-d", deviceHandle}
}

// ComputeCampaignSummary aggregates per-device statuses into a CampaignSummary.
func ComputeCampaignSummary(devices []models.CampaignDevice) models.CampaignSummary {
	summary := models.CampaignSummary{
		Total: len(devices),
	}
	for _, d := range devices {
		switch d.Status {
		case "pending", "in_progress":
			summary.Pending++
		case "success":
			summary.Success++
		case "deferred":
			summary.Deferred++
		case "skipped":
			summary.Skipped++
		case "wrong_fw":
			summary.WrongFW++
		case "failed":
			summary.Failed++
		}
	}
	return summary
}
