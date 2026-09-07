CREATE TABLE IF NOT EXISTS traffic_observations (
    id BIGSERIAL PRIMARY KEY,

    event_id TEXT NOT NULL UNIQUE,
    bus_id TEXT NOT NULL,

    event_timestamp TIMESTAMPTZ NOT NULL,

    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,

    location_bucket TEXT NOT NULL,
    time_bucket TEXT NOT NULL,

    motorcycle_count INTEGER NOT NULL DEFAULT 0,
    car_count INTEGER NOT NULL DEFAULT 0,
    truck_count INTEGER NOT NULL DEFAULT 0,
    bus_count INTEGER NOT NULL DEFAULT 0,

    vehicle_count INTEGER NOT NULL,
    congestion_score INTEGER NOT NULL,

    occupancy DOUBLE PRECISION,

    video_id TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS traffic_aggregates (
    id BIGSERIAL PRIMARY KEY,

    location_bucket TEXT NOT NULL,
    time_bucket TEXT NOT NULL,
    traffic_date DATE NOT NULL,

    observation_count BIGINT NOT NULL DEFAULT 0,

    sum_vehicle_count BIGINT NOT NULL DEFAULT 0,
    average_vehicle_count DOUBLE PRECISION NOT NULL DEFAULT 0,

    sum_congestion_score BIGINT NOT NULL DEFAULT 0,
    average_congestion_score DOUBLE PRECISION NOT NULL DEFAULT 0,

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (
        location_bucket,
        time_bucket,
        traffic_date
    )
);