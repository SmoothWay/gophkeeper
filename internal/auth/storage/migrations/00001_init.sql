-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users
(
    id            SERIAL PRIMARY KEY,
    login         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

CREATE TABLE IF NOT EXISTS apps
(
    id     SERIAL PRIMARY KEY,
    name   VARCHAR(255) NOT NULL UNIQUE,
    secret VARCHAR(255) NOT NULL UNIQUE
    );

INSERT INTO apps (id, name, secret)
VALUES (1, 'gophkeeper', 'test-secret')
    ON CONFLICT DO NOTHING;

-- +goose Down
DROP TABLE users;
DROP TABLE apps;
DROP EXTENSION IF EXISTS "uuid-ossp";