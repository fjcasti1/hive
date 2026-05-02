-- +goose Up
CREATE TABLE queue_new (
    id         INTEGER PRIMARY KEY,
    session    TEXT NOT NULL UNIQUE,
    message    TEXT NOT NULL DEFAULT '',
    pane       TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO queue_new
SELECT id, session, COALESCE(message, ''), '', created_at FROM queue;

DROP TABLE queue;
ALTER TABLE queue_new RENAME TO queue;

-- +goose Down
CREATE TABLE queue_old (
    id         INTEGER PRIMARY KEY,
    session    TEXT NOT NULL UNIQUE,
    message    TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO queue_old
SELECT id, session, message, created_at FROM queue;

DROP TABLE queue;
ALTER TABLE queue_old RENAME TO queue;
