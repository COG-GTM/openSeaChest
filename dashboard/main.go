// Package main is the entry point for the openSeaChest Fleet Health Dashboard.
//
// It starts the REST API server that exposes health monitoring endpoints
// backed by TimescaleDB for time-series storage of drive health metrics.
//
// Configuration is read from environment variables:
//   - DASHBOARD_LISTEN_ADDR: HTTP listen address (default ":8080")
//   - DASHBOARD_DATABASE_URL: TimescaleDB connection string
//     (default "postgres://dashboard:dashboard@localhost:5432/openseachest_health?sslmode=disable")
package main

import (
	"log"

	"github.com/COG-GTM/openSeaChest/dashboard/api"
)

func main() {
	cfg := api.DefaultConfig()

	log.Printf("openSeaChest Fleet Health Dashboard starting...")
	log.Printf("Listen address: %s", cfg.ListenAddr)

	if err := api.Run(cfg); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
