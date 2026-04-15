package agent

import (
	"fmt"
	"log"
	"time"

	"github.com/COG-GTM/openSeaChest/dashboard/models"
)

// DeviceTarget identifies a storage device for health collection.
type DeviceTarget struct {
	Handle string // OS device handle, e.g., /dev/sg0
	Serial string // Device serial number for metric tagging
}

// CollectionResult holds all metrics gathered from a single collection run.
type CollectionResult struct {
	Device      DeviceTarget
	SMARTCheck  *models.SMARTCheckResult
	Metrics     []models.HealthMetric
	Errors      []error
	CollectedAt time.Time
}

// CollectAll runs all health collectors (SMART check, SMART attributes, FARM logs)
// against a single device and aggregates the results.
func CollectAll(target DeviceTarget) *CollectionResult {
	result := &CollectionResult{
		Device:      target,
		CollectedAt: time.Now().UTC(),
	}

	// 1. Run SMART check
	smartResult, err := RunSMARTCheck(target.Handle)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("SMART check failed: %w", err))
		log.Printf("[WARN] SMART check failed for %s: %v", target.Serial, err)
	} else {
		smartResult.DeviceSerial = target.Serial
		result.SMARTCheck = smartResult
		result.Metrics = append(result.Metrics, SMARTCheckToMetrics(smartResult)...)
	}

	// 2. Collect SMART attributes
	smartAttrs, err := CollectSMARTAttributes(target.Handle, target.Serial)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("SMART attributes failed: %w", err))
		log.Printf("[WARN] SMART attributes collection failed for %s: %v", target.Serial, err)
	} else {
		result.Metrics = append(result.Metrics, smartAttrs...)
	}

	// 3. Collect FARM logs
	farmData, err := CollectFARMLogs(target.Handle)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("FARM log collection failed: %w", err))
		log.Printf("[WARN] FARM log collection failed for %s: %v", target.Serial, err)
	} else {
		result.Metrics = append(result.Metrics, FARMLogToMetrics(farmData)...)
	}

	return result
}
