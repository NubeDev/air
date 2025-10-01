#!/bin/bash
set -e

echo "🌱 Seeding TimescaleDB test database with fake data..."

# Check if container is running
if ! docker ps | grep -q air-timescale-test; then
    echo "❌ TimescaleDB test container is not running. Start it with: make db-test"
    exit 1
fi

# Copy SQL files to container
echo "� Copying SQL files to container..."
docker cp seed-test-data.sql air-timescale-test:/tmp/seed-test-data.sql
docker cp generate-history.sql air-timescale-test:/tmp/generate-history.sql

echo "�📊 Step 1: Inserting base test data (locations, devices, points, alerts)..."
docker exec air-timescale-test psql -U rubix -d rubix_test -f /tmp/seed-test-data.sql

echo "🕒 Step 2: Generating historical time-series data (this may take a few minutes)..."
docker exec air-timescale-test psql -U rubix -d rubix_test -f /tmp/generate-history.sql

echo "✅ Database seeding completed successfully!"
echo ""
echo "📈 Database Summary:"
docker exec air-timescale-test psql -U rubix -d rubix_test -c "
SELECT 'Locations: ' || count(*) FROM locations;
SELECT 'Groups: ' || count(*) FROM groups;
SELECT 'Hosts: ' || count(*) FROM hosts;
SELECT 'Networks: ' || count(*) FROM networks;
SELECT 'Devices: ' || count(*) FROM devices;
SELECT 'Points: ' || count(*) FROM points;
SELECT 'History records: ' || count(*) FROM histories;
SELECT 'Alerts: ' || count(*) FROM alerts;
"

echo ""
echo "🔗 Connection details:"
echo "  Host: localhost"
echo "  Port: 5434"
echo "  Database: rubix_test"
echo "  Username: rubix"
echo "  Password: rubix"
echo "  Connection string: postgres://rubix:rubix@localhost:5434/rubix_test"
echo ""
echo "🚀 Try some queries:"
echo "  # Recent temperature data"
echo "  docker exec air-timescale-test psql -U rubix -d rubix_test -c \"SELECT p.name, h.value, h.timestamp FROM histories h JOIN points p ON h.point_uuid = p.uuid WHERE p.name LIKE '%Temperature%' ORDER BY h.timestamp DESC LIMIT 10;\""
echo ""
echo "  # Active alerts"
echo "  docker exec air-timescale-test psql -U rubix -d rubix_test -c \"SELECT title, severity, status, created_at FROM alerts WHERE status = 'active';\""