CREATE TABLE vehicle_locations (
    id          UUID             PRIMARY KEY,
    vehicle_id  VARCHAR(32)      NOT NULL,
    latitude    DOUBLE PRECISION NOT NULL,
    longitude   DOUBLE PRECISION NOT NULL,
    timestamp   BIGINT           NOT NULL,
    created_at  TIMESTAMPTZ      NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX ux_vehicle_locations_vehicle_ts ON vehicle_locations (vehicle_id, timestamp);
