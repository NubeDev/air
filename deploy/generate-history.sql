-- Generate Historical Data for TimescaleDB Test Database
-- This script creates 30 days of realistic time-series data for all points

-- Function to generate realistic temperature variations
-- Office temperatures: 68-75°F with daily patterns
-- Warehouse temperatures: 70-85°F with less variation  
-- Supply air: 55-70°F
-- Return air: 70-80°F

DO $$
DECLARE
    point_rec RECORD;
    time_val TIMESTAMP;
    base_val FLOAT;
    variation FLOAT;
    daily_pattern FLOAT;
    noise FLOAT;
    hour_of_day INT;
    day_of_week INT;
    weekend_factor FLOAT;
BEGIN
    -- Generate data for the last 30 days
    FOR day_offset IN 0..29 LOOP
        FOR hour_offset IN 0..23 LOOP
            time_val := NOW() - INTERVAL '1 day' * day_offset - INTERVAL '1 hour' * hour_offset;
            hour_of_day := EXTRACT(hour FROM time_val);
            day_of_week := EXTRACT(dow FROM time_val); -- 0=Sunday, 6=Saturday
            
            -- Weekend factor (less activity on weekends)
            weekend_factor := CASE WHEN day_of_week IN (0, 6) THEN 0.8 ELSE 1.0 END;
            
            -- Daily pattern (warmer during day, cooler at night)
            daily_pattern := SIN((hour_of_day - 6) * PI() / 12) * 0.5;
            
            -- Generate data for each point
            FOR point_rec IN 
                SELECT uuid, name, present_value, unit, device_uuid
                FROM points 
                WHERE history_enable = true
            LOOP
                -- Base value from current present_value
                base_val := point_rec.present_value;
                
                -- Add realistic variations based on point type
                CASE 
                    -- Temperature points
                    WHEN point_rec.name LIKE '%Temperature%' THEN
                        variation := daily_pattern * 3 + (RANDOM() - 0.5) * 2; -- ±1°F noise, 3°F daily swing
                        
                    -- Humidity points  
                    WHEN point_rec.name LIKE '%Humidity%' THEN
                        variation := daily_pattern * -5 + (RANDOM() - 0.5) * 8; -- ±4% noise, 5% daily swing (inverse)
                        
                    -- Power/Energy points
                    WHEN point_rec.name LIKE '%Power%' THEN
                        -- Higher during business hours, lower on weekends
                        variation := daily_pattern * base_val * 0.3 * weekend_factor + (RANDOM() - 0.5) * base_val * 0.1;
                        
                    -- Energy meters (cumulative)
                    WHEN point_rec.name LIKE '%Energy%' AND point_rec.unit = 'kWh' THEN
                        -- Cumulative energy - always increasing
                        variation := -day_offset * 24 + RANDOM() * 10; -- Decrease as we go back in time
                        
                    -- Fan speeds, damper positions
                    WHEN point_rec.name LIKE '%Speed%' OR point_rec.name LIKE '%Position%' THEN
                        variation := daily_pattern * 15 + (RANDOM() - 0.5) * 10; -- ±5% noise, 15% daily swing
                        
                    -- Fan status (binary)
                    WHEN point_rec.name LIKE '%Status%' THEN
                        -- On during business hours mostly
                        variation := CASE 
                            WHEN hour_of_day BETWEEN 6 AND 22 AND day_of_week NOT IN (0, 6) THEN 1
                            WHEN hour_of_day BETWEEN 8 AND 20 AND day_of_week IN (0, 6) THEN 1
                            ELSE CASE WHEN RANDOM() > 0.8 THEN 1 ELSE 0 END
                        END - base_val;
                        
                    -- Default small variation
                    ELSE
                        variation := (RANDOM() - 0.5) * base_val * 0.05;
                END CASE;
                
                -- Insert into histories table
                INSERT INTO histories (id, point_uuid, host_uuid, value, timestamp)
                VALUES (
                    (SELECT COALESCE(MAX(history_id), 0) + 1 FROM histories),
                    point_rec.uuid,
                    (SELECT n.host_uuid FROM devices d JOIN networks n ON d.network_uuid = n.uuid WHERE d.uuid = point_rec.device_uuid),
                    GREATEST(0, base_val + variation),
                    time_val
                );
                
                -- Also insert into point_histories table
                INSERT INTO point_histories (point_uuid, value, timestamp)
                VALUES (
                    point_rec.uuid,
                    GREATEST(0, base_val + variation),
                    time_val
                );
                
            END LOOP;
        END LOOP;
        
        -- Show progress every 5 days
        IF day_offset % 5 = 0 THEN
            RAISE NOTICE 'Generated data for day % (% days ago)', day_offset + 1, day_offset;
        END IF;
    END LOOP;
    
    RAISE NOTICE 'Historical data generation completed!';
END $$;

-- Generate some metric logs (system performance data)
INSERT INTO metric_logs (name, value, timestamp)
SELECT 
    metric_name,
    base_value + (RANDOM() - 0.5) * variation_range,
    NOW() - INTERVAL '1 hour' * hour_offset
FROM (
    VALUES 
        ('cpu_usage', 45.0, 20.0),
        ('memory_usage', 65.0, 15.0),
        ('disk_usage', 78.0, 5.0),
        ('network_throughput', 125.0, 50.0),
        ('database_connections', 12.0, 8.0)
) AS metrics(metric_name, base_value, variation_range)
CROSS JOIN generate_series(0, 167) AS hour_offset; -- Last 7 days hourly

-- Update present_value in points table with latest values
UPDATE points 
SET present_value = latest.value
FROM (
    SELECT DISTINCT ON (point_uuid) 
        point_uuid, 
        value
    FROM histories 
    ORDER BY point_uuid, timestamp DESC
) latest
WHERE points.uuid = latest.point_uuid;

-- Summary statistics
SELECT 'Historical data generation summary:' as summary;
SELECT 'Total history records: ' || count(*) as history_count FROM histories;
SELECT 'Total point history records: ' || count(*) as point_history_count FROM point_histories;
SELECT 'Total metric logs: ' || count(*) as metric_count FROM metric_logs;
SELECT 'Date range: ' || MIN(timestamp) || ' to ' || MAX(timestamp) as date_range FROM histories;

-- Show some sample data
SELECT 'Sample temperature data from last 24 hours:' as sample_title;
SELECT 
    p.name,
    h.value,
    h.timestamp
FROM histories h
JOIN points p ON h.point_uuid = p.uuid
WHERE p.name LIKE '%Temperature%'
    AND h.timestamp > NOW() - INTERVAL '24 hours'
ORDER BY h.timestamp DESC
LIMIT 10;