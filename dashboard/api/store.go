// Package api implements the REST API for the openSeaChest Fleet Health Dashboard.
package api

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/COG-GTM/openSeaChest/dashboard/models"
)

// Store provides database operations for health metrics backed by TimescaleDB.
type Store struct {
	pool *pgxpool.Pool
}

// NewStore creates a new Store connected to the given TimescaleDB instance.
func NewStore(databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	return &Store{pool: pool}, nil
}

// Close shuts down the database connection pool.
func (s *Store) Close() {
	s.pool.Close()
}

// InsertMetrics writes a batch of health metrics into the health_metrics hypertable.
func (s *Store) InsertMetrics(ctx context.Context, metrics []models.HealthMetric) error {
	query := `
		INSERT INTO health_metrics (device_serial, time, metric_name, metric_value, unit, source)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	for _, m := range metrics {
		_, err := s.pool.Exec(ctx, query,
			m.DeviceSerial,
			m.Timestamp,
			m.MetricName,
			m.MetricValue,
			m.Unit,
			m.Source,
		)
		if err != nil {
			return fmt.Errorf("failed to insert metric %s for %s: %w", m.MetricName, m.DeviceSerial, err)
		}
	}
	return nil
}

// GetLatestMetrics retrieves the most recent metric values for a device.
func (s *Store) GetLatestMetrics(ctx context.Context, deviceSerial string) ([]models.HealthMetric, error) {
	query := `
		SELECT DISTINCT ON (metric_name)
			device_serial, time, metric_name, metric_value, unit, source
		FROM health_metrics
		WHERE device_serial = $1
		ORDER BY metric_name, time DESC
	`
	rows, err := s.pool.Query(ctx, query, deviceSerial)
	if err != nil {
		return nil, fmt.Errorf("failed to query latest metrics: %w", err)
	}
	defer rows.Close()

	var metrics []models.HealthMetric
	for rows.Next() {
		var m models.HealthMetric
		if err := rows.Scan(&m.DeviceSerial, &m.Timestamp, &m.MetricName, &m.MetricValue, &m.Unit, &m.Source); err != nil {
			return nil, fmt.Errorf("failed to scan metric row: %w", err)
		}
		metrics = append(metrics, m)
	}
	return metrics, rows.Err()
}

// GetHealthHistory retrieves time-series health metrics for a device within a time range.
// If metricName is non-empty, results are filtered to that specific metric.
func (s *Store) GetHealthHistory(ctx context.Context, deviceSerial string, from, to time.Time, metricName string) ([]models.HealthMetric, error) {
	var query string
	var args []interface{}

	if metricName != "" {
		query = `
			SELECT device_serial, time, metric_name, metric_value, unit, source
			FROM health_metrics
			WHERE device_serial = $1
			  AND time >= $2
			  AND time <= $3
			  AND metric_name = $4
			ORDER BY time ASC
		`
		args = []interface{}{deviceSerial, from, to, metricName}
	} else {
		query = `
			SELECT device_serial, time, metric_name, metric_value, unit, source
			FROM health_metrics
			WHERE device_serial = $1
			  AND time >= $2
			  AND time <= $3
			ORDER BY time ASC
		`
		args = []interface{}{deviceSerial, from, to}
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query health history: %w", err)
	}
	defer rows.Close()

	var metrics []models.HealthMetric
	for rows.Next() {
		var m models.HealthMetric
		if err := rows.Scan(&m.DeviceSerial, &m.Timestamp, &m.MetricName, &m.MetricValue, &m.Unit, &m.Source); err != nil {
			return nil, fmt.Errorf("failed to scan metric row: %w", err)
		}
		metrics = append(metrics, m)
	}
	return metrics, rows.Err()
}

// GetFleetHealthSummary returns aggregate health counts across all devices.
// It determines each device's status based on the most recent smart_check_status metric:
//   - 0.0 = healthy, 1.0 = warning, 2.0 = critical, anything else = unknown.
func (s *Store) GetFleetHealthSummary(ctx context.Context) (*models.FleetHealthSummary, error) {
	query := `
		SELECT device_serial, metric_value
		FROM (
			SELECT DISTINCT ON (device_serial)
				device_serial, metric_value, time
			FROM health_metrics
			WHERE metric_name = 'smart_check_status'
			ORDER BY device_serial, time DESC
		) latest
	`
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query fleet summary: %w", err)
	}
	defer rows.Close()

	summary := &models.FleetHealthSummary{
		Timestamp: time.Now().UTC(),
	}

	for rows.Next() {
		var serial string
		var statusValue float64
		if err := rows.Scan(&serial, &statusValue); err != nil {
			return nil, fmt.Errorf("failed to scan fleet row: %w", err)
		}

		status := models.HealthStatusUnknown
		switch statusValue {
		case 0.0:
			status = models.HealthStatusHealthy
			summary.HealthyCounts++
		case 1.0:
			status = models.HealthStatusWarning
			summary.WarningCounts++
		case 2.0:
			status = models.HealthStatusCritical
			summary.CriticalCounts++
		default:
			summary.UnknownCounts++
		}

		summary.Devices = append(summary.Devices, models.DeviceHealthSummary{
			DeviceSerial: serial,
			Status:       status,
		})
		summary.TotalDevices++
	}

	return summary, rows.Err()
}

// GetDeviceHealthSummary returns the current health summary for a single device.
func (s *Store) GetDeviceHealthSummary(ctx context.Context, deviceSerial string) (*models.DeviceHealthSummary, error) {
	metrics, err := s.GetLatestMetrics(ctx, deviceSerial)
	if err != nil {
		return nil, err
	}

	if len(metrics) == 0 {
		return &models.DeviceHealthSummary{
			DeviceSerial: deviceSerial,
			Status:       models.HealthStatusUnknown,
			LastChecked:  time.Time{},
			Message:      "no health data available for this device",
		}, nil
	}

	// Determine overall status from smart_check_status metric.
	status := models.HealthStatusUnknown
	var lastChecked time.Time
	for _, m := range metrics {
		if m.Timestamp.After(lastChecked) {
			lastChecked = m.Timestamp
		}
		if m.MetricName == "smart_check_status" {
			switch m.MetricValue {
			case 0.0:
				status = models.HealthStatusHealthy
			case 1.0:
				status = models.HealthStatusWarning
			case 2.0:
				status = models.HealthStatusCritical
			}
		}
	}

	return &models.DeviceHealthSummary{
		DeviceSerial: deviceSerial,
		Status:       status,
		LastChecked:  lastChecked,
		Metrics:      metrics,
	}, nil
}
