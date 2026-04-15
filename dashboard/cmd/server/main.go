// Package main is the entry point for the OpenSeaChest Fleet Health Dashboard
// alerting API server.
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/COG-GTM/openSeaChest/dashboard/api/handlers"
	"github.com/COG-GTM/openSeaChest/dashboard/api/services"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	store := services.NewStore()
	notifier := services.NewNotificationService()
	engine := services.NewRuleEngine(store, notifier)

	ruleHandler := handlers.NewRuleHandler(store)
	alertHandler := handlers.NewAlertHandler(store)

	mux := http.NewServeMux()

	// Alert Rule CRUD
	mux.HandleFunc("/api/v1/alerts/rules", ruleHandler.HandleRules)
	mux.HandleFunc("/api/v1/alerts/rules/", ruleHandler.HandleRuleByID)

	// Alert Endpoints
	mux.HandleFunc("/api/v1/alerts/active", alertHandler.HandleActiveAlerts)
	mux.HandleFunc("/api/v1/alerts/history", alertHandler.HandleAlertHistory)

	// Alert actions - these need special routing since Go's default mux
	// doesn't support path parameters natively.
	mux.HandleFunc("/api/v1/alerts/evaluate", func(w http.ResponseWriter, r *http.Request) {
		alertHandler.HandleEvaluate(w, r, engine)
	})

	// Catch-all for /api/v1/alerts/ paths that include an ID + action
	mux.HandleFunc("/api/v1/alerts/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case len(path) > len("/api/v1/alerts/") && containsSuffix(path, "/acknowledge"):
			alertHandler.HandleAcknowledge(w, r)
		case len(path) > len("/api/v1/alerts/") && containsSuffix(path, "/resolve"):
			alertHandler.HandleResolve(w, r)
		default:
			handlers.HandleNotFound(w, r)
		}
	})

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	log.Printf("OpenSeaChest Fleet Health Dashboard - Alerting API starting on :%s", port)
	if err := http.ListenAndServe(":"+port, withLogging(mux)); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// containsSuffix checks if a path ends with the given suffix.
func containsSuffix(path, suffix string) bool {
	return len(path) > len(suffix) && path[len(path)-len(suffix):] == suffix
}

// withLogging wraps an http.Handler with request logging middleware.
func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}
