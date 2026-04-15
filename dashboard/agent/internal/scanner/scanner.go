package scanner

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/COG-GTM/openSeaChest/dashboard/agent/internal/models"
)

// Scanner discovers storage devices by shelling out to openSeaChest CLI tools.
type Scanner struct {
	// BasicsPath is the filesystem path to the openSeaChest_Basics binary.
	BasicsPath string
}

// New creates a Scanner with the given binary path.
// If basicsPath is empty it defaults to "openSeaChest_Basics" (must be on PATH).
func New(basicsPath string) *Scanner {
	if basicsPath == "" {
		basicsPath = "openSeaChest_Basics"
	}
	return &Scanner{BasicsPath: basicsPath}
}

// ScanDevices runs `openSeaChest_Basics --scan` and returns the list of
// device handles found (e.g. /dev/sg0, /dev/sg1).
func (s *Scanner) ScanDevices() ([]string, error) {
	cmd := exec.Command(s.BasicsPath, "--scan")
	out, err := cmd.CombinedOutput()
	if err != nil {
		exitCode := extractExitCode(err)
		if exitCode == ExitNeedElevatedPrivileges {
			return nil, fmt.Errorf("elevated privileges required to scan devices: %w (exit code %d: %s)",
				err, exitCode, ExitCodeMessage(exitCode))
		}
		return nil, fmt.Errorf("scan failed: %w (exit code %d: %s)",
			err, exitCode, ExitCodeMessage(exitCode))
	}

	return parseScanOutput(string(out)), nil
}

// GetDeviceInfo runs `openSeaChest_Basics -d <handle> -i` and parses the
// output into a Device struct.
func (s *Scanner) GetDeviceInfo(handle string) (*models.Device, error) {
	cmd := exec.Command(s.BasicsPath, "-d", handle, "-i")
	out, err := cmd.CombinedOutput()
	if err != nil {
		exitCode := extractExitCode(err)
		if exitCode == ExitNeedElevatedPrivileges {
			return nil, fmt.Errorf("elevated privileges required for device %s: %w (exit code %d: %s)",
				handle, err, exitCode, ExitCodeMessage(exitCode))
		}
		return nil, fmt.Errorf("device info failed for %s: %w (exit code %d: %s)",
			handle, err, exitCode, ExitCodeMessage(exitCode))
	}

	dev := parseDeviceInfo(string(out))
	dev.DeviceHandle = handle
	return dev, nil
}

// DiscoverAll scans for devices, then collects info on each one.
// Devices that fail info collection are skipped (with errors logged to the
// returned error slice).
func (s *Scanner) DiscoverAll() ([]models.Device, []error) {
	handles, err := s.ScanDevices()
	if err != nil {
		return nil, []error{err}
	}

	var devices []models.Device
	var errs []error

	for _, h := range handles {
		dev, err := s.GetDeviceInfo(h)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		devices = append(devices, *dev)
	}

	return devices, errs
}

// ---------- stdout parsers ----------

// deviceHandleRe matches lines like:
//
//	/dev/sg0  SATA  ...
//	/dev/sg1  SAS   ...
var deviceHandleRe = regexp.MustCompile(`(?m)^\s*(/dev/\S+)\s+`)

func parseScanOutput(output string) []string {
	matches := deviceHandleRe.FindAllStringSubmatch(output, -1)
	seen := make(map[string]bool)
	var handles []string
	for _, m := range matches {
		h := m[1]
		if !seen[h] {
			seen[h] = true
			handles = append(handles, h)
		}
	}
	return handles
}

func parseDeviceInfo(output string) *models.Device {
	dev := &models.Device{}
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if k, v, ok := splitKV(line); ok {
			switch {
			case containsCI(k, "vendor") && containsCI(k, "id"):
				// skip vendor, we want model
			case containsCI(k, "product") && containsCI(k, "identification"):
				dev.Model = v
			case containsCI(k, "product") && containsCI(k, "revision"):
				dev.FirmwareRev = v
			case containsCI(k, "serial") && containsCI(k, "number"):
				dev.SerialNumber = v
			case containsCI(k, "interface"):
				dev.InterfaceType = v
			case containsCI(k, "world wide name") || containsCI(k, "wwn"):
				dev.WWN = v
			case containsCI(k, "capacity"):
				dev.CapacityBytes = parseCapacity(v)
			}
		}
	}
	return dev
}

// splitKV splits "Key: Value" lines. Returns key, value, ok.
func splitKV(line string) (string, string, bool) {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return "", "", false
	}
	return strings.TrimSpace(line[:idx]), strings.TrimSpace(line[idx+1:]), true
}

func containsCI(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// parseCapacity tries to extract a byte count from capacity strings.
// openSeaChest may print e.g. "500.11 GB [500,107,862,016 bytes]".
func parseCapacity(s string) int64 {
	re := regexp.MustCompile(`\[?\s*([\d,]+)\s*bytes\]?`)
	m := re.FindStringSubmatch(s)
	if m == nil {
		return 0
	}
	clean := strings.ReplaceAll(m[1], ",", "")
	var n int64
	fmt.Sscanf(clean, "%d", &n)
	return n
}

// ---------- helpers ----------

func extractExitCode(err error) int {
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode()
	}
	return -1
}
