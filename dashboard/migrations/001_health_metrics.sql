-- Migration: 001_health_metrics
-- Description: Create health_metrics hypertable for time-series storage of
--              drive health data collected by the openSeaChest Fleet Health Dashboard.
--
-- Prerequisites:
--   - TimescaleDB extension must be installed on the PostgreSQL instance.
--   - Run: CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Create the health_metrics table.
-- Each row represents a single metric observation for a specific device at a point in time.
CREATE TABLE IF NOT EXISTS health_metrics (
    -- Device identification
    device_serial TEXT        NOT NULL,

    -- Timestamp of the metric observation (used as the time dimension for the hypertable)
    time          TIMESTAMPTZ NOT NULL,

    -- Metric identification and value
    metric_name   TEXT        NOT NULL,
    metric_value  DOUBLE PRECISION NOT NULL,

    -- Optional metadata
    unit          TEXT        DEFAULT '',
    source        TEXT        NOT NULL DEFAULT 'unknown'
);

-- Convert the table into a TimescaleDB hypertable.
-- Partition by time with a default chunk interval of 7 days.
-- This optimizes time-range queries and enables automatic data retention policies.
SELECT create_hypertable(
    'health_metrics',
    'time',
    chunk_time_interval => INTERVAL '7 days',
    if_not_exists => TRUE
);

-- Index: Efficient lookup of all metrics for a specific device within a time range.
-- This is the primary access pattern for the device health history endpoint.
CREATE INDEX IF NOT EXISTS idx_health_metrics_device_time
    ON health_metrics (device_serial, time DESC);

-- Index: Efficient lookup of a specific metric across all devices.
-- Supports the fleet-wide summary endpoint and metric-filtered history queries.
CREATE INDEX IF NOT EXISTS idx_health_metrics_metric_name_time
    ON health_metrics (metric_name, time DESC);

-- Index: Composite index for the most common query pattern:
-- a specific metric for a specific device over a time range.
CREATE INDEX IF NOT EXISTS idx_health_metrics_device_metric_time
    ON health_metrics (device_serial, metric_name, time DESC);

-- Index: Source-based filtering for queries that need to isolate
-- metrics from a particular collector (smart_check, smart_attributes, farm_log).
CREATE INDEX IF NOT EXISTS idx_health_metrics_source_time
    ON health_metrics (source, time DESC);

-- Optional: Set up a data retention policy to automatically drop chunks
-- older than 90 days. Uncomment the following line to enable:
-- SELECT add_retention_policy('health_metrics', INTERVAL '90 days');

-- Optional: Set up continuous aggregates for hourly rollups.
-- Uncomment to create a materialized view with hourly averages:
--
-- CREATE MATERIALIZED VIEW health_metrics_hourly
-- WITH (timescaledb.continuous) AS
-- SELECT
--     device_serial,
--     metric_name,
--     time_bucket('1 hour', time) AS bucket,
--     AVG(metric_value) AS avg_value,
--     MIN(metric_value) AS min_value,
--     MAX(metric_value) AS max_value,
--     COUNT(*) AS sample_count
-- FROM health_metrics
-- GROUP BY device_serial, metric_name, bucket;
--
-- SELECT add_continuous_aggregate_policy('health_metrics_hourly',
--     start_offset => INTERVAL '2 hours',
--     end_offset   => INTERVAL '1 hour',
--     schedule_interval => INTERVAL '1 hour');
