package db

// queue.go holds the queue of pending agent messages: enqueuing (upsert),
// peeking the oldest entry, listing, and removing entries on delivery.
//
// An entry is keyed by agent_id — a stable identity for the agent (a Claude
// conversation id, an explicit id, or a tmux session name as a fallback). The
// label is what humans read; the locator describes how to reach the agent.

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	queueAddSQL = `
	INSERT INTO queue (agent_id, label, locator, message)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT(agent_id) DO UPDATE SET
		label      = excluded.label,
		locator    = excluded.locator,
		message    = excluded.message,
		created_at = CURRENT_TIMESTAMP
`

	queueListSQL = `
	SELECT
		id,
		agent_id,
		label,
		locator,
		message,
		created_at
	FROM queue
	ORDER BY created_at ASC, id ASC
`

	queueDeleteSQL = `
	DELETE FROM queue
	WHERE agent_id = $1
	RETURNING label, message, created_at
`
	queuePeekSQL = `
	SELECT
		id,
		agent_id,
		label,
		locator,
		message,
		created_at
	FROM queue
	ORDER BY created_at ASC, id ASC
	LIMIT 1
`
)

// queueEntry is a pending message awaiting delivery to an agent.
type queueEntry struct {
	ID        int64
	AgentID   string
	Label     string
	Locator   string
	Message   string
	CreatedAt time.Time
}

// Target returns the tmux pane to switch to for this entry, or "" if the agent
// has no navigable tmux locator (e.g. a bare process tracked only by cwd).
func (e queueEntry) Target() string {
	if pane, ok := strings.CutPrefix(e.Locator, "pane:"); ok {
		return pane
	}
	return ""
}

// Enqueue adds a message for the given agent. If the agent already has a queued
// message it is replaced and its position reset to the back of the queue.
func Enqueue(q Querier, agentID, label, locator, message string) error {
	_, err := q.Exec(queueAddSQL, agentID, label, locator, message)
	if err != nil {
		return fmt.Errorf("enqueue agent=%q: %w", agentID, err)
	}
	return nil
}

// Show returns the oldest queued entry without removing it, or nil if the queue
// is empty.
func Show(q Querier) (*queueEntry, error) {
	var (
		e         queueEntry
		createdAt string
	)
	err := q.QueryRow(queuePeekSQL).Scan(&e.ID, &e.AgentID, &e.Label, &e.Locator, &e.Message, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("peek: %w", err)
	}
	e.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &e, nil
}

// List returns all queued entries, oldest first.
func List(q Querier) ([]queueEntry, error) {
	rows, err := q.Query(queueListSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []queueEntry
	for rows.Next() {
		var (
			e         queueEntry
			createdAt string
		)
		if err := rows.Scan(&e.ID, &e.AgentID, &e.Label, &e.Locator, &e.Message, &createdAt); err != nil {
			return nil, err
		}
		e.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// deletedQueueEntry is the record returned when a queued entry is removed,
// carrying the fields needed to record it in history.
type deletedQueueEntry struct {
	Label      string
	Message    string
	NotifiedAt time.Time
}

// Delete removes the queued entry for the given agent and returns it, or nil
// if the agent had no queued entry.
func Delete(q Querier, agentID string) (*deletedQueueEntry, error) {
	var (
		e         deletedQueueEntry
		createdAt string
	)
	err := q.QueryRow(queueDeleteSQL, agentID).Scan(&e.Label, &e.Message, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("delete agent=%q: %w", agentID, err)
	}
	e.NotifiedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &e, nil
}
