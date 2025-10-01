-- Seed Test Data for Rubix OS TimescaleDB
-- This script populates the test database with realistic building automation data

-- Clear existing data (optional - uncomment if needed)
-- TRUNCATE TABLE histories, point_histories, alerts, points, devices, networks, hosts, groups, locations CASCADE;

-- Convert MySQL schema to PostgreSQL-compatible schema

-- Clear existing data first
TRUNCATE TABLE histories, point_histories, metric_logs CASCADE;
DELETE FROM alerts_tags;
DELETE FROM points_tags;
DELETE FROM devices_tags;
DELETE FROM networks_tags;
DELETE FROM alerts;
DELETE FROM points;
DELETE FROM devices;
DELETE FROM networks;
DELETE FROM hosts;
DELETE FROM groups;
DELETE FROM teams;
DELETE FROM members;
DELETE FROM plugins;
DELETE FROM tags;
DELETE FROM locations;

-- Insert Locations (Buildings)
INSERT INTO locations (uuid, name, description, address, city, state, zip, country, lat, lon, timezone) VALUES
('loc-office-a', 'Office Building A', 'Main corporate office building', '123 Business Ave', 'New York', 'NY', '10001', 'USA', '40.7505', '-73.9934', 'America/New_York'),
('loc-warehouse-b', 'Warehouse B', 'Distribution warehouse facility', '456 Industrial Blvd', 'Los Angeles', 'CA', '90001', 'USA', '34.0522', '-118.2437', 'America/Los_Angeles');

-- Insert Groups (Floors/Areas)
INSERT INTO groups (uuid, name, location_uuid, description) VALUES
('grp-office-floor1', 'Floor 1', 'loc-office-a', 'Ground floor - Lobby and reception'),
('grp-office-floor2', 'Floor 2', 'loc-office-a', 'Second floor - Offices and meeting rooms'),
('grp-warehouse-main', 'Main Floor', 'loc-warehouse-b', 'Main warehouse storage area'),
('grp-warehouse-office', 'Office Area', 'loc-warehouse-b', 'Administrative office area');

-- Insert Hosts (Controllers/Gateways)
INSERT INTO hosts (uuid, name, group_uuid, enable, description, ip, port, device_type, is_online, history_enable) VALUES
('host-office-ctrl1', 'Office Controller 1', 'grp-office-floor1', true, 'Main BAS controller for floor 1', '192.168.1.10', 502, 'Controller', true, true),
('host-office-ctrl2', 'Office Controller 2', 'grp-office-floor2', true, 'Main BAS controller for floor 2', '192.168.1.11', 502, 'Controller', true, true),
('host-warehouse-ctrl', 'Warehouse Controller', 'grp-warehouse-main', true, 'Main warehouse environmental controller', '192.168.2.10', 502, 'Controller', true, true),
('host-warehouse-gateway', 'IoT Gateway', 'grp-warehouse-office', true, 'LoRaWAN gateway for wireless sensors', '192.168.2.20', 1883, 'Gateway', true, true);

-- Insert Members (Users)
INSERT INTO members (uuid, first_name, last_name, username, email, permission, state) VALUES
('user-admin', 'John', 'Admin', 'admin', 'admin@company.com', 'admin', 'active'),
('user-operator', 'Sarah', 'Johnson', 'operator', 'sarah.johnson@company.com', 'operator', 'active'),
('user-technician', 'Mike', 'Smith', 'technician', 'mike.smith@company.com', 'technician', 'active');

-- Insert Teams
INSERT INTO teams (uuid, name, address, city, state, country) VALUES
('team-facilities', 'Facilities Management', '123 Business Ave', 'New York', 'NY', 'USA'),
('team-maintenance', 'Maintenance Team', '456 Industrial Blvd', 'Los Angeles', 'CA', 'USA');

-- Insert Tags
INSERT INTO tags (tag) VALUES
('HVAC'), ('Energy'), ('Temperature'), ('Humidity'), ('Critical'), ('Scheduled'), ('Emergency'), ('Sensor'), ('Actuator'), ('Meter');

-- Insert Plugins (Protocol drivers)
INSERT INTO plugins (uuid, name, enabled, message_level, message) VALUES
('plugin-modbus', 'Modbus TCP/IP', true, 'info', 'Modbus protocol driver'),
('plugin-bacnet', 'BACnet IP', true, 'info', 'BACnet protocol driver'),
('plugin-lorawan', 'LoRaWAN', true, 'info', 'LoRaWAN protocol driver'),
('plugin-mqtt', 'MQTT', true, 'info', 'MQTT protocol driver');

-- Insert Networks
INSERT INTO networks (uuid, name, description, enable, transport_type, plugin_uuid, host_uuid, ip, port, history_enable) VALUES
('net-modbus-office', 'Office Modbus Network', 'Modbus TCP network for office HVAC', true, 'tcp', 'plugin-modbus', 'host-office-ctrl1', '192.168.1.0', 502, true),
('net-bacnet-office', 'Office BACnet Network', 'BACnet IP network for office systems', true, 'ip', 'plugin-bacnet', 'host-office-ctrl2', '192.168.1.0', 47808, true),
('net-modbus-warehouse', 'Warehouse Modbus Network', 'Modbus TCP network for warehouse systems', true, 'tcp', 'plugin-modbus', 'host-warehouse-ctrl', '192.168.2.0', 502, true),
('net-lorawan-warehouse', 'Warehouse LoRaWAN', 'LoRaWAN network for wireless sensors', true, 'wireless', 'plugin-lorawan', 'host-warehouse-gateway', '192.168.2.20', 1883, true);

-- Insert Devices
INSERT INTO devices (uuid, name, description, enable, network_uuid, address_id, manufacture, model, history_enable) VALUES
-- Office Floor 1 Devices
('dev-ahu-1', 'AHU-1', 'Air Handling Unit Floor 1', true, 'net-modbus-office', 1, 'Johnson Controls', 'FX-PCD3633', true),
('dev-vav-101', 'VAV-101', 'Variable Air Volume Box Room 101', true, 'net-bacnet-office', 101, 'Trane', 'VAV-2000', true),
-- Office Floor 2 Devices  
('dev-ahu-2', 'AHU-2', 'Air Handling Unit Floor 2', true, 'net-modbus-office', 2, 'Johnson Controls', 'FX-PCD3633', true),
('dev-meter-main', 'Main Energy Meter', 'Main electrical meter', true, 'net-modbus-office', 10, 'Schneider', 'PM8000', true),
-- Warehouse Devices
('dev-rtu-1', 'RTU-1', 'Rooftop Unit 1', true, 'net-modbus-warehouse', 1, 'Carrier', 'RTU-50GS', true),
('dev-temp-sensor-1', 'Temp Sensor 1', 'Wireless temperature sensor zone 1', true, 'net-lorawan-warehouse', 1001, 'Sensirion', 'SHT85', true),
('dev-temp-sensor-2', 'Temp Sensor 2', 'Wireless temperature sensor zone 2', true, 'net-lorawan-warehouse', 1002, 'Sensirion', 'SHT85', true),
('dev-energy-meter-wh', 'Warehouse Energy Meter', 'Warehouse main electrical meter', true, 'net-modbus-warehouse', 20, 'Schneider', 'PM5000', true);

-- Insert Points
INSERT INTO points (uuid, name, description, enable, device_uuid, object_type, object_id, data_type, unit, history_enable, present_value) VALUES
-- AHU-1 Points
('pnt-ahu1-supply-temp', 'Supply Air Temperature', 'AHU-1 Supply air temperature', true, 'dev-ahu-1', 'analogInput', 1, 'real', '°F', true, 68.5),
('pnt-ahu1-return-temp', 'Return Air Temperature', 'AHU-1 Return air temperature', true, 'dev-ahu-1', 'analogInput', 2, 'real', '°F', true, 72.3),
('pnt-ahu1-fan-status', 'Supply Fan Status', 'AHU-1 Supply fan on/off status', true, 'dev-ahu-1', 'binaryInput', 1, 'boolean', '', true, 1),
('pnt-ahu1-fan-speed', 'Supply Fan Speed', 'AHU-1 Supply fan speed', true, 'dev-ahu-1', 'analogOutput', 1, 'real', '%', true, 75.0),

-- VAV-101 Points
('pnt-vav101-zone-temp', 'Zone Temperature', 'VAV-101 Zone temperature', true, 'dev-vav-101', 'analogInput', 1, 'real', '°F', true, 71.2),
('pnt-vav101-damper-pos', 'Damper Position', 'VAV-101 Damper position', true, 'dev-vav-101', 'analogOutput', 1, 'real', '%', true, 45.0),

-- AHU-2 Points  
('pnt-ahu2-supply-temp', 'Supply Air Temperature', 'AHU-2 Supply air temperature', true, 'dev-ahu-2', 'analogInput', 1, 'real', '°F', true, 69.1),
('pnt-ahu2-return-temp', 'Return Air Temperature', 'AHU-2 Return air temperature', true, 'dev-ahu-2', 'analogInput', 2, 'real', '°F', true, 73.8),
('pnt-ahu2-fan-status', 'Supply Fan Status', 'AHU-2 Supply fan on/off status', true, 'dev-ahu-2', 'binaryInput', 1, 'boolean', '', true, 1),

-- Main Energy Meter Points
('pnt-meter-main-power', 'Total Power', 'Main meter total power consumption', true, 'dev-meter-main', 'analogInput', 1, 'real', 'kW', true, 485.2),
('pnt-meter-main-energy', 'Total Energy', 'Main meter cumulative energy', true, 'dev-meter-main', 'analogInput', 2, 'real', 'kWh', true, 125430.0),

-- RTU-1 Points
('pnt-rtu1-discharge-temp', 'Discharge Air Temperature', 'RTU-1 Discharge air temperature', true, 'dev-rtu-1', 'analogInput', 1, 'real', '°F', true, 55.5),
('pnt-rtu1-return-temp', 'Return Air Temperature', 'RTU-1 Return air temperature', true, 'dev-rtu-1', 'analogInput', 2, 'real', '°F', true, 78.2),
('pnt-rtu1-cooling-cmd', 'Cooling Command', 'RTU-1 Cooling stage command', true, 'dev-rtu-1', 'analogOutput', 1, 'real', '%', true, 60.0),

-- Wireless Temperature Sensors
('pnt-temp1-temperature', 'Zone 1 Temperature', 'Wireless temperature sensor zone 1', true, 'dev-temp-sensor-1', 'analogInput', 1, 'real', '°F', true, 76.8),
('pnt-temp1-humidity', 'Zone 1 Humidity', 'Wireless humidity sensor zone 1', true, 'dev-temp-sensor-1', 'analogInput', 2, 'real', '%RH', true, 42.5),
('pnt-temp2-temperature', 'Zone 2 Temperature', 'Wireless temperature sensor zone 2', true, 'dev-temp-sensor-2', 'analogInput', 1, 'real', '°F', true, 77.5),
('pnt-temp2-humidity', 'Zone 2 Humidity', 'Wireless humidity sensor zone 2', true, 'dev-temp-sensor-2', 'analogInput', 2, 'real', '%RH', true, 38.9),

-- Warehouse Energy Meter
('pnt-meter-wh-power', 'Warehouse Power', 'Warehouse total power consumption', true, 'dev-energy-meter-wh', 'analogInput', 1, 'real', 'kW', true, 125.7),
('pnt-meter-wh-energy', 'Warehouse Energy', 'Warehouse cumulative energy', true, 'dev-energy-meter-wh', 'analogInput', 2, 'real', 'kWh', true, 45920.0);

-- Add some tags to networks, devices, and points
INSERT INTO networks_tags (network_uuid, tag_tag) VALUES
('net-modbus-office', 'HVAC'),
('net-bacnet-office', 'HVAC'), 
('net-modbus-warehouse', 'HVAC'),
('net-modbus-warehouse', 'Energy'),
('net-lorawan-warehouse', 'Sensor');

INSERT INTO devices_tags (device_uuid, tag_tag) VALUES
('dev-ahu-1', 'HVAC'),
('dev-ahu-1', 'Critical'),
('dev-ahu-2', 'HVAC'),
('dev-ahu-2', 'Critical'),
('dev-meter-main', 'Energy'),
('dev-meter-main', 'Critical'),
('dev-energy-meter-wh', 'Energy'),
('dev-rtu-1', 'HVAC'),
('dev-temp-sensor-1', 'Sensor'),
('dev-temp-sensor-2', 'Sensor');

INSERT INTO points_tags (point_uuid, tag_tag) VALUES
('pnt-ahu1-supply-temp', 'Temperature'),
('pnt-ahu1-return-temp', 'Temperature'),
('pnt-ahu2-supply-temp', 'Temperature'), 
('pnt-ahu2-return-temp', 'Temperature'),
('pnt-rtu1-discharge-temp', 'Temperature'),
('pnt-rtu1-return-temp', 'Temperature'),
('pnt-temp1-temperature', 'Temperature'),
('pnt-temp1-humidity', 'Humidity'),
('pnt-temp2-temperature', 'Temperature'),
('pnt-temp2-humidity', 'Humidity'),
('pnt-meter-main-power', 'Energy'),
('pnt-meter-main-energy', 'Energy'),
('pnt-meter-wh-power', 'Energy'),
('pnt-meter-wh-energy', 'Energy');

-- Insert some sample alerts
INSERT INTO alerts (uuid, host_uuid, entity_type, entity_uuid, type, status, severity, title, body, created_at, last_updated) VALUES
('alert-1', 'host-office-ctrl1', 'point', 'pnt-ahu1-supply-temp', 'temperature_high', 'active', 'warning', 'AHU-1 Supply Temperature High', 'Supply air temperature is above normal range (>75°F)', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '1 hour'),
('alert-2', 'host-warehouse-ctrl', 'device', 'dev-temp-sensor-1', 'communication_fault', 'active', 'critical', 'Temperature Sensor 1 Communication Fault', 'Lost communication with wireless temperature sensor in zone 1', NOW() - INTERVAL '4 hours', NOW() - INTERVAL '1 hour'),
('alert-3', 'host-office-ctrl2', 'point', 'pnt-meter-main-power', 'power_high', 'closed', 'warning', 'High Power Consumption', 'Building power consumption exceeded 500kW threshold', NOW() - INTERVAL '1 day', NOW() - INTERVAL '6 hours');

-- Add alert tags
INSERT INTO alerts_tags (alert_uuid, tag_tag) VALUES
('alert-1', 'HVAC'),
('alert-1', 'Temperature'),
('alert-2', 'Critical'),
('alert-2', 'Sensor'),
('alert-3', 'Energy');

SELECT 'Test data insertion completed successfully!' as status;
SELECT 'Networks: ' || count(*) FROM networks;
SELECT 'Devices: ' || count(*) FROM devices; 
SELECT 'Points: ' || count(*) FROM points;
SELECT 'Alerts: ' || count(*) FROM alerts;