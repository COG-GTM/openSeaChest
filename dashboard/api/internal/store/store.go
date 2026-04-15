// Package store provides PostgreSQL persistence for the Fleet Inventory API.
package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/COG-GTM/openSeaChest/dashboard/api/internal/models"
)

// Store wraps database operations.
type Store struct {
	DB *sql.DB
}

// New creates a Store with the given database connection.
func New(db *sql.DB) *Store {
	return &Store{DB: db}
}

// UpsertHost creates or updates a host record and returns its ID.
func (s *Store) UpsertHost(agentID, hostname string) (int64, error) {
	var id int64
	err := s.DB.QueryRow(`
		INSERT INTO hosts (agent_id, hostname, last_seen)
		VALUES ($1, $2, $3)
		ON CONFLICT (agent_id) DO UPDATE
			SET hostname  = EXCLUDED.hostname,
			    last_seen = EXCLUDED.last_seen
		RETURNING id
	`, agentID, hostname, time.Now().UTC()).Scan(&id)
	return id, err
}

// UpsertDevice creates or updates a device record.
func (s *Store) UpsertDevice(hostID int64, d models.DeviceReport) error {
	_, err := s.DB.Exec(`
		INSERT INTO devices (host_id, serial_number, model, firmware_rev,
		                     capacity_bytes, interface_type, wwn, last_seen)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (serial_number) DO UPDATE
			SET host_id        = EXCLUDED.host_id,
			    model          = EXCLUDED.model,
			    firmware_rev   = EXCLUDED.firmware_rev,
			    capacity_bytes = EXCLUDED.capacity_bytes,
			    interface_type = EXCLUDED.interface_type,
			    wwn            = EXCLUDED.wwn,
			    last_seen      = EXCLUDED.last_seen
	`, hostID, d.SerialNumber, d.Model, d.FirmwareRev,
		d.CapacityBytes, d.InterfaceType, d.WWN, time.Now().UTC())
	return err
}

// InsertSnapshot stores a point-in-time snapshot of raw device data.
func (s *Store) InsertSnapshot(deviceSerial string, rawData string) error {
	_, err := s.DB.Exec(`
		INSERT INTO device_snapshots (device_id, snapshot_time, raw_data)
		SELECT id, $1, $2 FROM devices WHERE serial_number = $3
	`, time.Now().UTC(), rawData, deviceSerial)
	return err
}

// DeviceFilter holds optional query filters for listing devices.
type DeviceFilter struct {
	Host        string
	Model       string
	Interface   string
	FirmwareRev string
	Page        int
	PerPage     int
}

// ListDevices returns a paginated, filtered list of devices.
func (s *Store) ListDevices(f DeviceFilter) (*models.PaginatedDevices, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PerPage < 1 || f.PerPage > 100 {
		f.PerPage = 20
	}

	where, args := buildWhereClause(f)

	// Count total
	countQuery := "SELECT COUNT(*) FROM devices d JOIN hosts h ON d.host_id = h.id" + where
	var total int
	if err := s.DB.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count devices: %w", err)
	}

	// Fetch page
	offset := (f.Page - 1) * f.PerPage
	dataQuery := fmt.Sprintf(`
		SELECT d.id, d.host_id, h.hostname, d.serial_number, d.model,
		       d.firmware_rev, d.capacity_bytes, d.interface_type, d.wwn,
		       d.first_seen, d.last_seen
		FROM devices d
		JOIN hosts h ON d.host_id = h.id
		%s
		ORDER BY d.last_seen DESC
		LIMIT %d OFFSET %d
	`, where, f.PerPage, offset)

	rows, err := s.DB.Query(dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}
	defer rows.Close()

	var devices []models.DeviceResponse
	for rows.Next() {
		var d models.DeviceResponse
		if err := rows.Scan(&d.ID, &d.HostID, &d.Hostname, &d.SerialNumber,
			&d.Model, &d.FirmwareRev, &d.CapacityBytes, &d.InterfaceType,
			&d.WWN, &d.FirstSeen, &d.LastSeen); err != nil {
			return nil, fmt.Errorf("scan device row: %w", err)
		}
		devices = append(devices, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate device rows: %w", err)
	}

	return &models.PaginatedDevices{
		Devices: devices,
		Total:   total,
		Page:    f.Page,
		PerPage: f.PerPage,
	}, nil
}

// GetDeviceBySerial returns a single device by serial number.
func (s *Store) GetDeviceBySerial(serial string) (*models.DeviceResponse, error) {
	var d models.DeviceResponse
	err := s.DB.QueryRow(`
		SELECT d.id, d.host_id, h.hostname, d.serial_number, d.model,
		       d.firmware_rev, d.capacity_bytes, d.interface_type, d.wwn,
		       d.first_seen, d.last_seen
		FROM devices d
		JOIN hosts h ON d.host_id = h.id
		WHERE d.serial_number = $1
	`, serial).Scan(&d.ID, &d.HostID, &d.Hostname, &d.SerialNumber,
		&d.Model, &d.FirmwareRev, &d.CapacityBytes, &d.InterfaceType,
		&d.WWN, &d.FirstSeen, &d.LastSeen)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get device by serial: %w", err)
	}
	return &d, nil
}

// ListHosts returns all known hosts.
func (s *Store) ListHosts() ([]models.HostResponse, error) {
	rows, err := s.DB.Query(`
		SELECT id, agent_id, hostname, last_seen
		FROM hosts
		ORDER BY last_seen DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list hosts: %w", err)
	}
	defer rows.Close()

	var hosts []models.HostResponse
	for rows.Next() {
		var h models.HostResponse
		if err := rows.Scan(&h.ID, &h.AgentID, &h.Hostname, &h.LastSeen); err != nil {
			return nil, fmt.Errorf("scan host row: %w", err)
		}
		hosts = append(hosts, h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate host rows: %w", err)
	}
	return hosts, nil
}

// ---------- helpers ----------

func buildWhereClause(f DeviceFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	idx := 1

	if f.Host != "" {
		conditions = append(conditions, fmt.Sprintf("h.hostname ILIKE $%d", idx))
		args = append(args, "%"+f.Host+"%")
		idx++
	}
	if f.Model != "" {
		conditions = append(conditions, fmt.Sprintf("d.model ILIKE $%d", idx))
		args = append(args, "%"+f.Model+"%")
		idx++
	}
	if f.Interface != "" {
		conditions = append(conditions, fmt.Sprintf("d.interface_type ILIKE $%d", idx))
		args = append(args, "%"+f.Interface+"%")
		idx++
	}
	if f.FirmwareRev != "" {
		conditions = append(conditions, fmt.Sprintf("d.firmware_rev ILIKE $%d", idx))
		args = append(args, "%"+f.FirmwareRev+"%")
		idx++
	}

	if len(conditions) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}
