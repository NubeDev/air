#!/bin/bash
set -e

echo "Initializing Rubix OS schema for TimescaleDB..."

# Create the schema with PostgreSQL-compatible syntax
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-'EOSQL'
-- Convert MySQL schema to PostgreSQL-compatible schema

CREATE TABLE IF NOT EXISTS "alerts" (
    "uuid" text PRIMARY KEY,
    "host_uuid" text,
    "entity_type" text,
    "entity_uuid" text,
    "type" text,
    "status" text,
    "severity" text,
    "target" text,
    "title" text,
    "body" text,
    "notified" boolean DEFAULT false,
    "notified_at" timestamp,
    "created_at" timestamp,
    "last_updated" timestamp,
    "source" text,
    "emailed" boolean DEFAULT false,
    "emailed_at" timestamp,
    "acknowledge_timeout" timestamp
);

CREATE TABLE IF NOT EXISTS "tags" (
    "tag" text NOT NULL UNIQUE PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS "alerts_tags" (
    "tag_tag" text NOT NULL,
    "alert_uuid" text,
    PRIMARY KEY ("tag_tag", "alert_uuid"),
    CONSTRAINT "fk_alerts_tags_alert" FOREIGN KEY ("alert_uuid") REFERENCES "alerts"("uuid") ON DELETE CASCADE,
    CONSTRAINT "fk_alerts_tags_tag" FOREIGN KEY ("tag_tag") REFERENCES "tags"("tag") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "alert_closeds" (
    "uuid" text PRIMARY KEY,
    "host_uuid" text,
    "entity_type" text,
    "entity_uuid" text,
    "type" text,
    "status" text,
    "severity" text,
    "target" text,
    "title" text,
    "body" text,
    "notified" boolean DEFAULT false,
    "notified_at" timestamp,
    "created_at" timestamp,
    "last_updated" timestamp,
    "closed_at" timestamp
);

CREATE TABLE IF NOT EXISTS "teams" (
    "uuid" text PRIMARY KEY,
    "name" text NOT NULL UNIQUE,
    "address" text,
    "city" text,
    "state" text,
    "zip" integer,
    "country" text,
    "lat" text,
    "lon" text,
    "time_zone" text
);

CREATE TABLE IF NOT EXISTS "alert_teams" (
    "alert_uuid" text NOT NULL,
    "team_uuid" text NOT NULL,
    PRIMARY KEY ("alert_uuid", "team_uuid"),
    CONSTRAINT "fk_alerts_teams" FOREIGN KEY ("alert_uuid") REFERENCES "alerts"("uuid") ON DELETE CASCADE,
    CONSTRAINT "fk_teams_alerts" FOREIGN KEY ("team_uuid") REFERENCES "teams"("uuid") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "locations" (
    "uuid" text UNIQUE PRIMARY KEY,
    "name" text NOT NULL UNIQUE,
    "description" text,
    "address" text,
    "city" text,
    "state" text,
    "zip" text,
    "country" text,
    "lat" text,
    "lon" text,
    "timezone" text
);

CREATE TABLE IF NOT EXISTS "groups" (
    "uuid" text PRIMARY KEY,
    "name" text NOT NULL,
    "location_uuid" text NOT NULL,
    "description" text,
    CONSTRAINT "fk_locations_groups" FOREIGN KEY ("location_uuid") REFERENCES "locations"("uuid") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "hosts" (
    "uuid" text PRIMARY KEY,
    "global_uuid" text,
    "device_type" text,
    "name" text NOT NULL,
    "group_uuid" text NOT NULL,
    "enable" boolean,
    "description" text,
    "ip" text,
    "bios_port" integer,
    "port" integer,
    "http_s" boolean,
    "is_online" boolean,
    "is_valid_token" boolean,
    "external_token" text,
    "virtual_ip" text,
    "received_bytes" integer,
    "sent_bytes" integer,
    "connected_since" text,
    "history_enable" boolean DEFAULT false,
    "ping_fail_count" integer,
    "ros_version" text,
    "ros_restart_expression" text,
    "timezone" text,
    "disconnected_since" text,
    "reboot_expression" text,
    CONSTRAINT "fk_groups_hosts" FOREIGN KEY ("group_uuid") REFERENCES "groups"("uuid") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "plugins" (
    "uuid" text UNIQUE PRIMARY KEY,
    "name" text NOT NULL UNIQUE,
    "enabled" boolean,
    "config" bytea,
    "storage" bytea,
    "message_level" text,
    "message" text
);

CREATE TABLE IF NOT EXISTS "networks" (
    "uuid" text UNIQUE PRIMARY KEY,
    "name" text,
    "description" text,
    "enable" boolean DEFAULT false,
    "in_fault" boolean,
    "message_level" text,
    "message_code" text,
    "message" text,
    "message_fail" text,
    "last_ok" timestamp,
    "last_fail" timestamp,
    "created_at" timestamp,
    "updated_at" timestamp,
    "last_write" timestamp,
    "manufacture" text,
    "model" text,
    "writeable_network" boolean,
    "thing_class" text,
    "thing_ref" text,
    "thing_type" text,
    "transport_type" text NOT NULL,
    "plugin_uuid" text NOT NULL,
    "plugin_name" text,
    "number_of_networks_permitted" integer,
    "network_interface" text,
    "ip" text,
    "port" integer,
    "network_mask" integer,
    "address_id" text,
    "address_uuid" text,
    "serial_port" text,
    "serial_baud_rate" integer,
    "serial_stop_bits" integer,
    "serial_parity" text,
    "serial_data_bits" integer,
    "serial_timeout" integer,
    "serial_connected" boolean,
    "host" text,
    "max_poll_rate" double precision,
    "has_polling_statistics" boolean,
    "global_uuid" text,
    "connection" text DEFAULT 'Connected',
    "connection_message" text,
    "source_uuid" text,
    "source_plugin_name" text,
    "is_clone" boolean DEFAULT false,
    "host_uuid" text,
    "history_enable" boolean DEFAULT false,
    "supports_device_ping" boolean,
    "config" jsonb,
    CONSTRAINT "fk_hosts_networks" FOREIGN KEY ("host_uuid") REFERENCES "hosts"("uuid") ON DELETE CASCADE,
    CONSTRAINT "fk_plugins_network" FOREIGN KEY ("plugin_uuid") REFERENCES "plugins"("uuid") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "devices" (
    "uuid" text UNIQUE PRIMARY KEY,
    "name" text,
    "description" text,
    "enable" boolean DEFAULT false,
    "in_fault" boolean,
    "message_level" text,
    "message_code" text,
    "message" text,
    "message_fail" text,
    "last_ok" timestamp,
    "last_fail" timestamp,
    "created_at" timestamp,
    "updated_at" timestamp,
    "last_write" timestamp,
    "thing_class" text,
    "thing_ref" text,
    "thing_type" text,
    "manufacture" text,
    "model" text,
    "address_id" integer,
    "zero_mode" boolean,
    "poll_delay_points_ms" integer,
    "address_uuid" text,
    "host" text,
    "port" integer,
    "device_mac" integer,
    "device_object_id" integer,
    "network_number" integer,
    "max_adpu" integer,
    "segmentation" text,
    "device_mask" integer,
    "type_serial" boolean,
    "transport_type" text,
    "supports_rpm" boolean,
    "supports_wpm" boolean,
    "network_uuid" text NOT NULL,
    "number_of_devices_permitted" integer,
    "fast_poll_rate" double precision,
    "normal_poll_rate" double precision,
    "slow_poll_rate" double precision,
    "delay_between_points" integer,
    "device_timeout" integer,
    "connection" text DEFAULT 'Connected',
    "connection_message" text,
    "source_uuid" text,
    "history_enable" boolean DEFAULT false,
    "delay_between_points_ms" integer,
    "config" jsonb,
    "is_clone" boolean DEFAULT false,
    "disable_grouping" boolean DEFAULT false,
    "enable_concurrency" boolean DEFAULT false,
    "concurrency_limit" integer DEFAULT 10,
    CONSTRAINT "fk_networks_devices" FOREIGN KEY ("network_uuid") REFERENCES "networks"("uuid") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "points" (
    "uuid" text UNIQUE PRIMARY KEY,
    "name" text,
    "description" text,
    "enable" boolean DEFAULT false,
    "created_at" timestamp,
    "updated_at" timestamp,
    "last_write" timestamp,
    "thing_class" text,
    "thing_ref" text,
    "thing_type" text,
    "in_fault" boolean,
    "message_level" text,
    "message_code" text,
    "message" text,
    "message_fail" text,
    "last_ok" timestamp,
    "last_fail" timestamp,
    "present_value" double precision,
    "original_value" double precision,
    "display_value" text,
    "write_value" double precision,
    "write_value_original" double precision,
    "current_priority" integer,
    "write_priority" integer,
    "is_output" boolean,
    "is_type_bool" boolean,
    "in_sync" boolean,
    "fallback" double precision,
    "device_uuid" text NOT NULL,
    "enable_writeable" boolean,
    "math_on_present_value" text,
    "math_on_write_value" text,
    "cov" double precision,
    "object_type" text,
    "object_id" integer,
    "data_type" text,
    "is_bitwise" boolean,
    "bitwise_index" integer,
    "object_encoding" text,
    "io_number" text,
    "io_type" text,
    "address_id" integer,
    "address_length" integer,
    "address_uuid" text,
    "next_available_address" boolean,
    "decimal" integer,
    "multiplication_factor" double precision,
    "scale_enable" boolean,
    "scale_in_min" double precision,
    "scale_in_max" double precision,
    "scale_out_min" double precision,
    "scale_out_max" double precision,
    "offset" double precision,
    "unit_type" text,
    "unit" text,
    "unit_to" text,
    "value_updated_flag" boolean,
    "point_priority_array_mode" text,
    "write_mode" text,
    "read_write_type" text,
    "write_poll_required" boolean,
    "read_poll_required" boolean,
    "poll_priority" text,
    "poll_rate" text,
    "ba_cnet_write_to_pv" boolean,
    "history_enable" boolean DEFAULT false,
    "history_type" text,
    "history_interval" integer,
    "history_cov_threshold" double precision,
    "connection" text DEFAULT 'Connected',
    "connection_message" text,
    "source_uuid" text,
    "last_history_value" double precision,
    "point_state" text,
    "last_history_timestamp" timestamp,
    "config" jsonb,
    "poll_on_startup" boolean,
    "is_clone" boolean DEFAULT false,
    CONSTRAINT "fk_devices_points" FOREIGN KEY ("device_uuid") REFERENCES "devices"("uuid") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "histories" (
    "history_id" serial PRIMARY KEY,
    "id" integer,
    "point_uuid" text,
    "host_uuid" text,
    "value" double precision,
    "timestamp" timestamp
);

CREATE TABLE IF NOT EXISTS "point_histories" (
    "id" serial PRIMARY KEY,
    "point_uuid" text NOT NULL,
    "value" double precision,
    "timestamp" timestamp
);

CREATE TABLE IF NOT EXISTS "history_postgres_logs" (
    "id" serial PRIMARY KEY,
    "point_uuid" text,
    "host_uuid" text,
    "value" double precision,
    "timestamp" timestamp
);

CREATE TABLE IF NOT EXISTS "metric_logs" (
    "id" serial PRIMARY KEY,
    "name" text,
    "value" double precision,
    "timestamp" timestamp
);

CREATE TABLE IF NOT EXISTS "history_logs" (
    "id" serial PRIMARY KEY,
    "host_uuid" text,
    "last_sync_id" integer,
    "timestamp" timestamp,
    "metric_log_last_sync_id" integer,
    "metric_log_timestamp" timestamp
);

CREATE TABLE IF NOT EXISTS "members" (
    "uuid" text UNIQUE PRIMARY KEY,
    "first_name" text,
    "last_name" text,
    "username" text NOT NULL UNIQUE,
    "password" text,
    "email" text,
    "permission" text,
    "state" text
);

CREATE TABLE IF NOT EXISTS "schedules" (
    "uuid" text UNIQUE PRIMARY KEY,
    "name" text NOT NULL UNIQUE,
    "enable" boolean DEFAULT false,
    "thing_class" text,
    "thing_type" text,
    "is_active" boolean,
    "active_weekly" boolean,
    "active_exception" boolean,
    "active_event" boolean,
    "enable_payload" boolean,
    "min_payload" double precision,
    "max_payload" double precision,
    "payload" double precision,
    "default_payload" double precision,
    "period_start" integer,
    "period_stop" integer,
    "next_start" integer,
    "next_stop" integer,
    "period_start_string" text,
    "period_stop_string" text,
    "next_start_string" text,
    "next_stop_string" text,
    "schedule" jsonb,
    "created_at" timestamp,
    "updated_at" timestamp,
    "last_write" timestamp,
    "global_uuid" text,
    "connection" text DEFAULT 'Connected',
    "connection_message" text,
    "nullable_output" boolean
);

-- Create additional support tables
CREATE TABLE IF NOT EXISTS "priorities" (
    "point_uuid" text NOT NULL PRIMARY KEY,
    "p1" double precision,
    "p2" double precision,
    "p3" double precision,
    "p4" double precision,
    "p5" double precision,
    "p6" double precision,
    "p7" double precision,
    "p8" double precision,
    "p9" double precision,
    "p10" double precision,
    "p11" double precision,
    "p12" double precision,
    "p13" double precision,
    "p14" double precision,
    "p15" double precision,
    "p16" double precision,
    CONSTRAINT "fk_points_priority" FOREIGN KEY ("point_uuid") REFERENCES "points"("uuid") ON DELETE CASCADE
);

-- Create some essential indexes
CREATE INDEX IF NOT EXISTS "idx_alerts_status" ON "alerts"("status");
CREATE INDEX IF NOT EXISTS "idx_history_logs_id" ON "history_logs"("id");
CREATE INDEX IF NOT EXISTS "idx_metric_logs_id" ON "metric_logs"("id");
CREATE INDEX IF NOT EXISTS "idx_point_histories_id" ON "point_histories"("id");
CREATE UNIQUE INDEX IF NOT EXISTS "idx_histories_point_uuid_host_uuid_timestamp" ON "histories"("point_uuid","host_uuid","timestamp");
CREATE UNIQUE INDEX IF NOT EXISTS "idx_point_histories_point_uuid_timestamp" ON "point_histories"("point_uuid","timestamp");
CREATE UNIQUE INDEX IF NOT EXISTS "idx_hosts_name_group_uuid" ON "hosts"("name","group_uuid");
CREATE UNIQUE INDEX IF NOT EXISTS "idx_groups_name_location_uuid" ON "groups"("name","location_uuid");
CREATE UNIQUE INDEX IF NOT EXISTS "idx_networks_name_host_uuid" ON "networks"("name",COALESCE("host_uuid",''));
CREATE UNIQUE INDEX IF NOT EXISTS "idx_devices_name_network_uuid" ON "devices"("name","network_uuid");
CREATE UNIQUE INDEX IF NOT EXISTS "idx_points_name_device_uuid" ON "points"("name","device_uuid");

-- Create tag-related tables
CREATE TABLE IF NOT EXISTS "networks_tags" (
    "tag_tag" text NOT NULL,
    "network_uuid" text,
    PRIMARY KEY ("tag_tag", "network_uuid"),
    CONSTRAINT "fk_networks_tags_tag" FOREIGN KEY ("tag_tag") REFERENCES "tags"("tag") ON DELETE CASCADE,
    CONSTRAINT "fk_networks_tags_network" FOREIGN KEY ("network_uuid") REFERENCES "networks"("uuid") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "devices_tags" (
    "tag_tag" text NOT NULL,
    "device_uuid" text,
    PRIMARY KEY ("tag_tag", "device_uuid"),
    CONSTRAINT "fk_devices_tags_device" FOREIGN KEY ("device_uuid") REFERENCES "devices"("uuid") ON DELETE CASCADE,
    CONSTRAINT "fk_devices_tags_tag" FOREIGN KEY ("tag_tag") REFERENCES "tags"("tag") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "points_tags" (
    "tag_tag" text NOT NULL,
    "point_uuid" text,
    PRIMARY KEY ("tag_tag", "point_uuid"),
    CONSTRAINT "fk_points_tags_tag" FOREIGN KEY ("tag_tag") REFERENCES "tags"("tag") ON DELETE CASCADE,
    CONSTRAINT "fk_points_tags_point" FOREIGN KEY ("point_uuid") REFERENCES "points"("uuid") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "hosts_tags" (
    "host_uuid" text,
    "tag_tag" text NOT NULL,
    PRIMARY KEY ("host_uuid", "tag_tag"),
    CONSTRAINT "fk_hosts_tags_host" FOREIGN KEY ("host_uuid") REFERENCES "hosts"("uuid") ON DELETE CASCADE,
    CONSTRAINT "fk_hosts_tags_tag" FOREIGN KEY ("tag_tag") REFERENCES "tags"("tag") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "teams_tags" (
    "team_uuid" text,
    "tag_tag" text NOT NULL,
    PRIMARY KEY ("team_uuid", "tag_tag"),
    CONSTRAINT "fk_teams_tags_team" FOREIGN KEY ("team_uuid") REFERENCES "teams"("uuid") ON DELETE CASCADE,
    CONSTRAINT "fk_teams_tags_tag" FOREIGN KEY ("tag_tag") REFERENCES "tags"("tag") ON DELETE CASCADE
);

-- Create meta tag tables
CREATE TABLE IF NOT EXISTS "network_meta_tags" (
    "network_uuid" text NOT NULL,
    "key" text,
    "value" text,
    PRIMARY KEY ("network_uuid", "key"),
    CONSTRAINT "fk_networks_meta_tags" FOREIGN KEY ("network_uuid") REFERENCES "networks"("uuid") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "device_meta_tags" (
    "device_uuid" text NOT NULL,
    "key" text,
    "value" text,
    PRIMARY KEY ("device_uuid", "key"),
    CONSTRAINT "fk_devices_meta_tags" FOREIGN KEY ("device_uuid") REFERENCES "devices"("uuid") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "point_meta_tags" (
    "point_uuid" text NOT NULL,
    "key" text,
    "value" text,
    PRIMARY KEY ("point_uuid", "key"),
    CONSTRAINT "fk_points_meta_tags" FOREIGN KEY ("point_uuid") REFERENCES "points"("uuid") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "alert_meta_tags" (
    "alert_uuid" text NOT NULL,
    "key" text,
    "value" text,
    PRIMARY KEY ("alert_uuid", "key"),
    CONSTRAINT "fk_alerts_meta_tags" FOREIGN KEY ("alert_uuid") REFERENCES "alerts"("uuid") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "team_meta_tags" (
    "team_uuid" text NOT NULL,
    "key" text,
    "value" text,
    PRIMARY KEY ("team_uuid", "key"),
    CONSTRAINT "fk_teams_meta_tags" FOREIGN KEY ("team_uuid") REFERENCES "teams"("uuid") ON DELETE CASCADE
);

EOSQL

echo "Rubix OS schema initialized successfully!"