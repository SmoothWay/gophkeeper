-- +goose Up
CREATE TABLE IF NOT EXISTS credentials
(
    id                 INTEGER PRIMARY KEY,
    tag                TEXT,
    login              TEXT NOT NULL UNIQUE,
    password           TEXT,
    comment            TEXT,
    created_at         INTEGER
);

CREATE TABLE IF NOT EXISTS text
(
    id                 INTEGER PRIMARY KEY,
    tag                TEXT,
    key                TEXT NOT NULL UNIQUE,
    value              TEXT,
    comment            TEXT,
    created_at         INTEGER
);

CREATE TABLE IF NOT EXISTS binary
(
    id                 INTEGER PRIMARY KEY,
    tag                TEXT,
    key                TEXT NOT NULL UNIQUE,
    value              BLOB,
    comment            TEXT,
    created_at         INTEGER
);

CREATE TABLE IF NOT EXISTS card
(
    id                 INTEGER PRIMARY KEY,
    tag                TEXT,
    number             TEXT NOT NULL UNIQUE,
    exp                TEXT,
    cvv                INTEGER,
    comment            TEXT,
    created_at         INTEGER
);


-- +goose Down
DROP TABLE credentials;
DROP TABLE text;
DROP TABLE binary;
DROP TABLE card;