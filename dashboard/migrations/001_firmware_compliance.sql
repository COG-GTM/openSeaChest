-- 001_firmware_compliance.sql
-- Database migration for firmware compliance auditing tables.
--
-- Tables:
--   firmware_policies    — defines firmware requirements per device model pattern
--   compliance_results   — stores per-device compliance evaluation results
--   firmware_campaigns   — tracks firmware update rollout campaigns
--   campaign_devices     — per-device status within a campaign

CREATE TABLE IF NOT EXISTS firmware_policies (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    name            TEXT    NOT NULL,
    model_pattern   TEXT    NOT NULL,   -- glob pattern matched against device product_identification
    required_firmware TEXT  NOT NULL,   -- firmware revision that devices must be running
    severity        TEXT    NOT NULL DEFAULT 'warning', -- 'critical', 'warning', 'info'
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS compliance_results (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    device_id         TEXT    NOT NULL,
    policy_id         INTEGER NOT NULL REFERENCES firmware_policies(id),
    compliant         BOOLEAN NOT NULL,
    current_firmware  TEXT    NOT NULL,
    checked_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(device_id, policy_id)
);

CREATE TABLE IF NOT EXISTS firmware_campaigns (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT    NOT NULL,
    policy_id     INTEGER NOT NULL REFERENCES firmware_policies(id),
    firmware_file TEXT    NOT NULL,
    status        TEXT    NOT NULL DEFAULT 'created', -- 'created', 'in_progress', 'completed', 'cancelled'
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at    DATETIME,
    completed_at  DATETIME
);

CREATE TABLE IF NOT EXISTS campaign_devices (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    campaign_id   INTEGER NOT NULL REFERENCES firmware_campaigns(id),
    device_serial TEXT    NOT NULL,
    status        TEXT    NOT NULL DEFAULT 'pending', -- 'pending','in_progress','success','deferred','skipped','wrong_fw','failed'
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_compliance_results_device   ON compliance_results(device_id);
CREATE INDEX IF NOT EXISTS idx_compliance_results_policy   ON compliance_results(policy_id);
CREATE INDEX IF NOT EXISTS idx_campaign_devices_campaign    ON campaign_devices(campaign_id);
CREATE INDEX IF NOT EXISTS idx_campaign_devices_serial      ON campaign_devices(device_serial);
