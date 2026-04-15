// Package main is the entry point for the openSeaChest Fleet Health Dashboard
// firmware compliance auditing server.
//
// It initialises an SQLite database, runs the migration, and starts an HTTP
// server exposing the firmware compliance API.
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"github.com/COG-GTM/openSeaChest/dashboard/api"
)

func main() {
	dbPath := os.Getenv("DASHBOARD_DB_PATH")
	if dbPath == "" {
		dbPath = "dashboard.db"
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := runMigrations(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	router := api.NewRouter(db)

	addr := os.Getenv("DASHBOARD_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	fmt.Printf("openSeaChest Fleet Health Dashboard listening on %s\n", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// runMigrations reads and executes the SQL migration file.
func runMigrations(db *sql.DB) error {
	migration, err := os.ReadFile("migrations/001_firmware_compliance.sql")
	if err != nil {
		return fmt.Errorf("reading migration file: %w", err)
	}

	_, err = db.Exec(string(migration))
	if err != nil {
		return fmt.Errorf("executing migration: %w", err)
	}

	log.Println("database migrations applied successfully")
	return nil
}
