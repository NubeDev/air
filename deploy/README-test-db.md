# TimescaleDB Test Database Setup

This directory contains the setup for a TimescaleDB test database with the Rubix OS schema pre-installed.

## Files

- `Dockerfile.timescale-test`: Docker image definition for the test database
- `init-schema.sh`: PostgreSQL-compatible version of the Rubix OS schema
- `enable-timescaledb.sql`: TimescaleDB configuration and hypertable setup

## Usage

### Start the test database
```bash
make db-test
```

### Connect to the database
```bash
# Using psql
psql postgres://rubix:rubix@localhost:5434/rubix_test

# Using Docker
docker exec -it air-timescale-test psql -U rubix -d rubix_test
```

### Stop the test database
```bash
make down-test
```

### Remove test database completely (including data)
```bash
docker compose -f deploy/docker-compose.yml down timescale-test -v
```

## Database Details

- **Host**: localhost
- **Port**: 5434
- **Database**: rubix_test
- **Username**: rubix
- **Password**: rubix
- **Connection String**: `postgres://rubix:rubix@localhost:5434/rubix_test`

## Schema Features

The test database includes:

1. **Complete Rubix OS Schema**: All tables from the original MySQL schema converted to PostgreSQL
2. **TimescaleDB Extensions**: Time-series tables converted to hypertables for better performance
3. **Optimized Indexes**: Indexes optimized for time-series queries
4. **Compression**: Automatic compression for chunks older than 7 days
5. **Proper Constraints**: Foreign key relationships and constraints maintained

## Hypertables

The following tables are configured as TimescaleDB hypertables:
- `histories` - Historical data points
- `point_histories` - Point-specific historical data
- `history_postgres_logs` - PostgreSQL sync logs
- `metric_logs` - System metrics

## Testing

This database is ideal for:
- Testing time-series data ingestion
- Performance testing with large datasets
- Testing queries against historical data
- Development and debugging of Rubix OS features