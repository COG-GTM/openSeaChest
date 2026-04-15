-- 001_inventory.sql
-- Fleet Health Dashboard — initial schema for device inventory.

BEGIN;

-- hosts: each unique agent/host that reports in.
CREATE TABLE IF NOT EXISTS hosts (
    id          BIGSERIAL PRIMARY KEY,
    agent_id    TEXT NOT NULL UNIQUE,
    hostname    TEXT NOT NULL,
    first_seen  TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_hosts_agent_id  ON hosts (agent_id);
CREATE INDEX IF NOT EXISTS idx_hosts_hostname  ON hosts (hostname);

-- devices: one row per physical storage device (keyed by serial number).
CREATE TABLE IF NOT EXISTS devices (
    id              BIGSERIAL PRIMARY KEY,
    host_id         BIGINT NOT NULL REFERENCES hosts(id) ON DELETE CASCADE,
    serial_number   TEXT NOT NULL UNIQUE,
    model           TEXT NOT NULL DEFAULT '',
    firmware_rev    TEXT NOT NULL DEFAULT '',
    capacity_bytes  BIGINT NOT NULL DEFAULT 0,
    interface_type  TEXT NOT NULL DEFAULT '',
    wwn             TEXT NOT NULL DEFAULT '',
    first_seen      TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_devices_host_id        ON devices (host_id);
CREATE INDEX IF NOT EXISTS idx_devices_serial_number  ON devices (serial_number);
CREATE INDEX IF NOT EXISTS idx_devices_model          ON devices (model);
CREATE INDEX IF NOT EXISTS idx_devices_interface_type  ON devices (interface_type);
CREATE INDEX IF NOT EXISTS idx_devices_firmware_rev    ON devices (firmware_rev);

-- device_snapshots: point-in-time captures of raw device state.
CREATE TABLE IF NOT EXISTS device_snapshots (
    id              BIGSERIAL PRIMARY KEY,
    device_id       BIGINT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    snapshot_time   TIMESTAMPTZ NOT NULL DEFAULT now(),
    raw_data        JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_snapshots_device_id     ON device_snapshots (device_id);
CREATE INDEX IF NOT EXISTS idx_snapshots_snapshot_time ON device_snapshots (snapshot_time);

COMMIT;
