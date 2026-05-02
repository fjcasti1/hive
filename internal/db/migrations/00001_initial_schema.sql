-- +goose Up
CREATE TABLE queue (
    id            INTEGER PRIMARY KEY,
    session       TEXT NOT NULL UNIQUE,
    message       TEXT,
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE queue;
