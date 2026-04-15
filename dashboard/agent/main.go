// osc-agent is the per-host agent that discovers local storage devices via
// openSeaChest CLI tools and reports them to the central Fleet Inventory API.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/COG-GTM/openSeaChest/dashboard/agent/internal/models"
	"github.com/COG-GTM/openSeaChest/dashboard/agent/internal/reporter"
	"github.com/COG-GTM/openSeaChest/dashboard/agent/internal/scanner"
)

func main() {
	var (
		apiURL     = flag.String("api-url", "http://localhost:8080", "Base URL of the Fleet Inventory API")
		agentID    = flag.String("agent-id", "", "Unique identifier for this agent (defaults to hostname)")
		basicsPath = flag.String("basics-path", "", "Path to openSeaChest_Basics binary (default: search PATH)")
		interval   = flag.Duration("interval", 0, "Run continuously at this interval (0 = run once)")
	)
	flag.Parse()

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	if *agentID == "" {
		*agentID = hostname
	}

	sc := scanner.New(*basicsPath)
	rp := reporter.New(*apiURL)

	run := func() {
		log.Println("Starting device discovery...")
		devices, errs := sc.DiscoverAll()
		for _, e := range errs {
			log.Printf("WARNING: %v", e)
		}

		log.Printf("Discovered %d device(s)", len(devices))
		for i, d := range devices {
			log.Printf("  [%d] %s  model=%s  serial=%s  fw=%s  iface=%s",
				i+1, d.DeviceHandle, d.Model, d.SerialNumber, d.FirmwareRev, d.InterfaceType)
		}

		report := &models.InventoryReport{
			AgentID:  *agentID,
			Hostname: hostname,
			Devices:  devices,
		}

		if err := rp.SendReport(report); err != nil {
			log.Printf("ERROR: Failed to send report: %v", err)
			return
		}
		log.Println("Report sent successfully")
	}

	if *interval <= 0 {
		run()
		return
	}

	fmt.Printf("osc-agent running every %s  agent_id=%s  api=%s\n", *interval, *agentID, *apiURL)
	ticker := time.NewTicker(*interval)
	defer ticker.Stop()
	run() // run immediately, then on tick
	for range ticker.C {
		run()
	}
}
