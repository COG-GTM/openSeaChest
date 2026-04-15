// Package api wires up the HTTP router for the firmware compliance auditing
// API. All endpoints live under /api/v1/firmware/.
package api

import (
	"database/sql"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/COG-GTM/openSeaChest/dashboard/api/handlers"
)

// NewRouter returns a chi.Router with all firmware compliance endpoints registered.
func NewRouter(db *sql.DB) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.SetHeader("Content-Type", "application/json"))

	policyH := handlers.NewPolicyHandler(db)
	complianceH := handlers.NewComplianceHandler(db)
	campaignH := handlers.NewCampaignHandler(db)

	r.Route("/api/v1/firmware", func(r chi.Router) {
		// Firmware Policy CRUD
		r.Post("/policies", policyH.CreatePolicy)
		r.Get("/policies", policyH.ListPolicies)
		r.Put("/policies/{id}", policyH.UpdatePolicy)

		// Fleet Compliance Report
		r.Get("/compliance", complianceH.GetFleetCompliance)

		// Firmware Update Campaigns
		r.Post("/campaigns", campaignH.CreateCampaign)
		r.Get("/campaigns/{id}", campaignH.GetCampaign)
		r.Post("/campaigns/{id}/results", campaignH.ReportDeviceResult)
	})

	return r
}
