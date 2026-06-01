-- +goose Up
-- Decouple a queue entry's identity from the tmux session name. An "agent" is
-- now keyed by a stable agent_id (a Claude conversation id, an explicit --id, or
-- a tmux session as a fallback), with a human-readable label for display and a
-- locator describing how to reach it (e.g. "pane:%5" for tmux, "cwd:/path" for a
-- bare process). This lets agents that live outside tmux get distinct, navigable
-- entries instead of colliding on a single session name.
CREATE TABLE queue_new (
    id         INTEGER PRIMARY KEY,
    agent_id   TEXT NOT NULL UNIQUE,
    label      TEXT NOT NULL DEFAULT '',
    locator    TEXT NOT NULL DEFAULT '',
    message    TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO queue_new (id, agent_id, label, locator, message, created_at)
SELECT
    id,
    session,
    session,
    CASE WHEN pane != '' THEN 'pane:' || pane ELSE '' END,
    message,
    created_at
FROM queue;

DROP TABLE queue;
ALTER TABLE queue_new RENAME TO queue;

-- +goose Down
CREATE TABLE queue_old (
    id         INTEGER PRIMARY KEY,
    session    TEXT NOT NULL UNIQUE,
    message    TEXT NOT NULL DEFAULT '',
    pane       TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO queue_old (id, session, message, pane, created_at)
SELECT
    id,
    agent_id,
    message,
    CASE WHEN locator LIKE 'pane:%' THEN substr(locator, 6) ELSE '' END,
    created_at
FROM queue;

DROP TABLE queue;
ALTER TABLE queue_old RENAME TO queue;
