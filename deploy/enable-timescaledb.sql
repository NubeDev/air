-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Convert time-series tables to hypertables
-- Only converting tables that have timestamp columns and are likely time-series data

-- Create hypertables for time-series data
SELECT create_hypertable('histories', 'timestamp', 
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

SELECT create_hypertable('point_histories', 'timestamp', 
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

SELECT create_hypertable('history_postgres_logs', 'timestamp', 
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

SELECT create_hypertable('metric_logs', 'timestamp', 
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- Create compression policy for better performance (optional but recommended)
-- Compress chunks older than 7 days
SELECT add_compression_policy('histories', INTERVAL '7 days');
SELECT add_compression_policy('point_histories', INTERVAL '7 days');
SELECT add_compression_policy('history_postgres_logs', INTERVAL '7 days');
SELECT add_compression_policy('metric_logs', INTERVAL '7 days');

-- Create retention policy to automatically drop data older than 1 year (optional)
-- Uncomment the following lines if you want automatic data retention
-- SELECT add_retention_policy('histories', INTERVAL '1 year');
-- SELECT add_retention_policy('point_histories', INTERVAL '1 year');
-- SELECT add_retention_policy('history_postgres_logs', INTERVAL '1 year');
-- SELECT add_retention_policy('metric_logs', INTERVAL '1 year');

-- Create some time-series specific indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_histories_timestamp ON histories (timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_point_histories_timestamp ON point_histories (timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_history_postgres_logs_timestamp ON history_postgres_logs (timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_metric_logs_timestamp ON metric_logs (timestamp DESC);

-- Create indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_histories_point_timestamp ON histories (point_uuid, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_point_histories_point_timestamp ON point_histories (point_uuid, timestamp DESC);

SELECT 'TimescaleDB configuration completed successfully!' as result;