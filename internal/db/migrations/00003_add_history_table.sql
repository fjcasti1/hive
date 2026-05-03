-- +goose Up
CREATE TABLE history (
    id          INTEGER PRIMARY KEY,
    session     TEXT NOT NULL,
    message     TEXT NOT NULL DEFAULT '',
    notified_at DATETIME NOT NULL,
    resolved_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE history;
