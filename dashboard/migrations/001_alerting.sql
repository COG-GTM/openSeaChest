-- Migration: 001_alerting.sql
-- Description: Create tables for the alerting and notification engine.
-- Part of: WI-4 — Alerting & Notification Engine
--
-- This migration creates the core schema for the Fleet Health Dashboard
-- alerting system, including alert rules, alerts, alert history tracking,
-- and notification channels.

BEGIN;

-- =============================================================================
-- alert_rules: Defines conditions under which alerts should fire.
-- =============================================================================
CREATE TABLE IF NOT EXISTS alert_rules (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    rule_type   TEXT NOT NULL CHECK (rule_type IN (
                    'smart_trip',
                    'smart_warning',
                    'temperature_threshold',
                    'firmware_non_compliance',
                    'agent_stale',
                    'reallocated_sector_growth'
                )),
    config      JSONB NOT NULL DEFAULT '{}',
    severity    TEXT NOT NULL DEFAULT 'warning' CHECK (severity IN ('critical', 'warning', 'info')),
    enabled     BOOLEAN NOT NULL DEFAULT TRUE,
    notification_channels UUID[] DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_alert_rules_rule_type ON alert_rules (rule_type);
CREATE INDEX idx_alert_rules_enabled ON alert_rules (enabled);

COMMENT ON TABLE alert_rules IS 'Alert rules define conditions that trigger alerts on storage devices.';
COMMENT ON COLUMN alert_rules.rule_type IS 'Built-in rule type: smart_trip, smart_warning, temperature_threshold, firmware_non_compliance, agent_stale, reallocated_sector_growth';
COMMENT ON COLUMN alert_rules.config IS 'Rule-specific configuration in JSONB format (e.g., threshold values, firmware versions)';
COMMENT ON COLUMN alert_rules.notification_channels IS 'Array of notification channel IDs to notify when this rule fires';

-- =============================================================================
-- alerts: Active and historical alert instances.
-- =============================================================================
CREATE TABLE IF NOT EXISTS alerts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id         UUID NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
    device_serial   TEXT NOT NULL,
    state           TEXT NOT NULL DEFAULT 'firing' CHECK (state IN ('ok', 'firing', 'acknowledged', 'resolved')),
    message         TEXT NOT NULL DEFAULT '',
    first_seen      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    acknowledged_at TIMESTAMPTZ,
    resolved_at     TIMESTAMPTZ,
    acknowledged_by TEXT,
    resolved_by     TEXT
);

CREATE INDEX idx_alerts_state ON alerts (state);
CREATE INDEX idx_alerts_rule_id ON alerts (rule_id);
CREATE INDEX idx_alerts_device_serial ON alerts (device_serial);
CREATE INDEX idx_alerts_active ON alerts (state) WHERE state IN ('firing', 'acknowledged');

-- Unique constraint for deduplication: only one active alert per rule+device
CREATE UNIQUE INDEX idx_alerts_dedup ON alerts (rule_id, device_serial) WHERE state IN ('firing', 'acknowledged');

COMMENT ON TABLE alerts IS 'Alert instances tracking device conditions. State machine: OK -> FIRING -> ACKNOWLEDGED -> RESOLVED.';
COMMENT ON COLUMN alerts.state IS 'Alert state: ok (initial), firing (condition triggered), acknowledged (user ack), resolved (condition cleared or manual)';
COMMENT ON COLUMN alerts.last_seen IS 'Updated on deduplication — when the same condition fires again on the same device';

-- =============================================================================
-- alert_history: Audit trail of alert state transitions.
-- =============================================================================
CREATE TABLE IF NOT EXISTS alert_history (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alert_id    UUID NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    old_state   TEXT NOT NULL CHECK (old_state IN ('ok', 'firing', 'acknowledged', 'resolved')),
    new_state   TEXT NOT NULL CHECK (new_state IN ('ok', 'firing', 'acknowledged', 'resolved')),
    changed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    changed_by  TEXT NOT NULL DEFAULT 'system'
);

CREATE INDEX idx_alert_history_alert_id ON alert_history (alert_id);
CREATE INDEX idx_alert_history_changed_at ON alert_history (changed_at);

COMMENT ON TABLE alert_history IS 'Audit log of all alert state transitions for compliance and debugging.';

-- =============================================================================
-- notification_channels: Configured delivery channels for alert notifications.
-- =============================================================================
CREATE TABLE IF NOT EXISTS notification_channels (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         TEXT NOT NULL,
    channel_type TEXT NOT NULL CHECK (channel_type IN ('webhook', 'email')),
    config       JSONB NOT NULL DEFAULT '{}',
    enabled      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notification_channels_type ON notification_channels (channel_type);
CREATE INDEX idx_notification_channels_enabled ON notification_channels (enabled);

COMMENT ON TABLE notification_channels IS 'Notification delivery channels. Supports webhook (with retry/backoff) and email (SMTP stub).';
COMMENT ON COLUMN notification_channels.config IS 'Channel-specific config: webhook={url, headers, secret}, email={smtp_host, smtp_port, from_address, to_addresses, ...}';

-- =============================================================================
-- Helper function: auto-update updated_at on alert_rules changes
-- =============================================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_alert_rules_updated_at
    BEFORE UPDATE ON alert_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMIT;
