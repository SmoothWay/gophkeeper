-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS store
(
    id                 SERIAL PRIMARY KEY,
    user_id            SERIAL,
    type               VARCHAR,
    key                VARCHAR,
    data               BYTEA,
    created_at_client  BIGINT,
    created_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE store;
DROP EXTENSION IF EXISTS "uuid-ossp"