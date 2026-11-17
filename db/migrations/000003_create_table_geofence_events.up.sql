CREATE TABLE geofence_events (
    id          UUID             PRIMARY KEY,
    vehicle_id  VARCHAR(32)      NOT NULL,
    geofence_id UUID             NOT NULL REFERENCES geofences (id),
    event       VARCHAR(32)      NOT NULL,
    latitude    DOUBLE PRECISION NOT NULL,
    longitude   DOUBLE PRECISION NOT NULL,
    timestamp   BIGINT           NOT NULL,
    received_at TIMESTAMPTZ      NOT NULL DEFAULT now(),
    UNIQUE (vehicle_id, geofence_id, event, timestamp)
);
