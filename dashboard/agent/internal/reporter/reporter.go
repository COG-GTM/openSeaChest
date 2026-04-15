// Package reporter sends device inventory data to the central Fleet API.
package reporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/COG-GTM/openSeaChest/dashboard/agent/internal/models"
)

// Reporter posts inventory reports to the central API.
type Reporter struct {
	// BaseURL is the root URL of the fleet inventory API, e.g. "http://fleet-api:8080".
	BaseURL string
	// Client is the HTTP client used for requests.
	Client *http.Client
}

// New creates a Reporter targeting the given API base URL.
func New(baseURL string) *Reporter {
	return &Reporter{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendReport POSTs the inventory report to POST /api/v1/agents/{agent_id}/report.
func (r *Reporter) SendReport(report *models.InventoryReport) error {
	body, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}

	reportURL := fmt.Sprintf("%s/api/v1/agents/%s/report", r.BaseURL, url.PathEscape(report.AgentID))
	req, err := http.NewRequest(http.MethodPost, reportURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.Client.Do(req)
	if err != nil {
		return fmt.Errorf("send report: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API returned HTTP %d", resp.StatusCode)
	}
	return nil
}
