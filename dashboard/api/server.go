package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

// ServerConfig holds configuration for the API server.
type ServerConfig struct {
	ListenAddr  string // e.g., ":8080"
	DatabaseURL string // TimescaleDB connection string
}

// DefaultConfig returns a ServerConfig populated from environment variables
// with sensible defaults.
func DefaultConfig() ServerConfig {
	addr := os.Getenv("DASHBOARD_LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	dbURL := os.Getenv("DASHBOARD_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://dashboard:dashboard@localhost:5432/openseachest_health?sslmode=disable"
	}

	return ServerConfig{
		ListenAddr:  addr,
		DatabaseURL: dbURL,
	}
}

// Run starts the API server with graceful shutdown support.
func Run(cfg ServerConfig) error {
	store, err := NewStore(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to initialize store: %w", err)
	}
	defer store.Close()

	handler := NewHandler(store)
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	}).Methods("GET")

	handler.RegisterRoutes(router)

	srv := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	errCh := make(chan error, 1)
	go func() {
		log.Printf("Starting API server on %s", cfg.ListenAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Printf("Received signal %v, shutting down...", sig)
	case err := <-errCh:
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return srv.Shutdown(ctx)
}
