// Package scanner wraps openSeaChest CLI tools to discover storage devices.
package scanner

// Exit codes from openSeaChest utilities.
// Sourced from include/openseachest_util_options.h — eUtilExitCodes enum.
const (
	ExitNoError                  = 0
	ExitErrorInCommandLine       = 1
	ExitInvalidDeviceHandle      = 2
	ExitOperationFailure         = 3
	ExitOperationNotSupported    = 4
	ExitOperationAborted         = 5
	ExitPathNotFound             = 6
	ExitCannotOpenFile           = 7
	ExitFileAlreadyExists        = 8
	ExitNeedElevatedPrivileges   = 9
	ExitNotEnoughResources       = 10
	ExitErrorWritingFile         = 11
	ExitNoDevice                 = 12
	ExitDeviceBusy               = 13
	ExitInsecurePath             = 14
	ExitToolSpecificStartingCode = 32
)

// ExitCodeMessage returns a human-readable description for an openSeaChest exit code.
func ExitCodeMessage(code int) string {
	switch code {
	case ExitNoError:
		return "no error"
	case ExitErrorInCommandLine:
		return "error in command line"
	case ExitInvalidDeviceHandle:
		return "invalid device handle"
	case ExitOperationFailure:
		return "operation failure"
	case ExitOperationNotSupported:
		return "operation not supported"
	case ExitOperationAborted:
		return "operation aborted"
	case ExitPathNotFound:
		return "path not found"
	case ExitCannotOpenFile:
		return "cannot open file"
	case ExitFileAlreadyExists:
		return "file already exists"
	case ExitNeedElevatedPrivileges:
		return "need elevated privileges (run as root/administrator)"
	case ExitNotEnoughResources:
		return "not enough resources"
	case ExitErrorWritingFile:
		return "error writing file"
	case ExitNoDevice:
		return "no device found"
	case ExitDeviceBusy:
		return "device busy"
	case ExitInsecurePath:
		return "insecure path"
	default:
		if code >= ExitToolSpecificStartingCode {
			return "tool-specific error"
		}
		return "unknown error"
	}
}
