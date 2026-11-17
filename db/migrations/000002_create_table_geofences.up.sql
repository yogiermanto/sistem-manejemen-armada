CREATE TABLE geofences (
    id          UUID             PRIMARY KEY,
    name        VARCHAR(100)     NOT NULL,
    latitude    DOUBLE PRECISION NOT NULL,
    longitude   DOUBLE PRECISION NOT NULL,
    radius_m    DOUBLE PRECISION NOT NULL CHECK (radius_m > 0),
    created_at  TIMESTAMPTZ      NOT NULL DEFAULT now()
);

INSERT INTO geofences (id, name, latitude, longitude, radius_m) VALUES (gen_random_uuid(), 'Default', -6.2088, 106.8456, 50);
