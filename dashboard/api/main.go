// Fleet Inventory API — central REST service for the openSeaChest Fleet Health Dashboard.
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/COG-GTM/openSeaChest/dashboard/api/internal/handlers"
	"github.com/COG-GTM/openSeaChest/dashboard/api/internal/middleware"
	"github.com/COG-GTM/openSeaChest/dashboard/api/internal/store"

	_ "github.com/lib/pq"
)

func main() {
	var (
		listenAddr = flag.String("addr", ":8080", "Listen address")
		dbDSN      = flag.String("dsn", envOrDefault("DATABASE_URL",
			"postgres://fleet:fleet@localhost:5432/fleet_inventory?sslmode=disable"),
			"PostgreSQL connection string")
	)
	flag.Parse()

	db, err := sql.Open("postgres", *dbDSN)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	s := store.New(db)
	h := handlers.New(s)

	mux := http.NewServeMux()

	// POST /api/v1/agents/{agent_id}/report
	mux.HandleFunc("/api/v1/agents/", h.ReceiveReport)

	// GET /api/v1/devices and GET /api/v1/devices/{serial}
	mux.HandleFunc("/api/v1/devices", h.ListDevices)
	mux.HandleFunc("/api/v1/devices/", h.GetDevice)

	// GET /api/v1/hosts
	mux.HandleFunc("/api/v1/hosts", h.ListHosts)

	// Health check
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	handler := middleware.Logging(middleware.JSON(mux))

	log.Printf("Fleet Inventory API listening on %s", *listenAddr)
	if err := http.ListenAndServe(*listenAddr, handler); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
